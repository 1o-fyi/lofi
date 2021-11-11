package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"syscall"

	"filippo.io/age"
	"git.sr.ht/~lofi/lib"
	"github.com/keep-network/keep-core/pkg/bls"
	"github.com/spf13/cobra"
)

const (
	ErrIncorrectFlag = "\nsetup: incorrect or missing flag(s)\n"
)

var (
	flagApi           = "https://1o.fyi"
	flagMsg           = ""
	flagRecip         = ""
	flagPath          = ""
	flagUser          = ""
	flagMsgId         = ""
	flagMinimalOutput bool

	RootCmd *cobra.Command = &cobra.Command{
		Use:   "lofi",
		Short: __SHORT,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			os.Stdout.Write([]byte("\n"))
		},
	}

	infoCmd *cobra.Command = &cobra.Command{
		Use:     "info",
		Aliases: []string{"i"},
		Short:   "learn about lofi cli",
		Run: func(cmd *cobra.Command, args []string) {
			os.Stdout.Write([]byte(__INFO))
		},
	}

	// this command prints out a formatted registry string from a private key
	// it requires the computation of G2.
	mapCmd *cobra.Command = &cobra.Command{
		Use:     "fmt",
		Aliases: []string{"f"},
		Short:   `formats public keys for a registry line | username::age_public_key::G2_public_key`,
		Run: func(cmd *cobra.Command, args []string) {
			if anyInvalid(flagPath, flagUser) {
				os.Stdout.Write([]byte("setup: requires a -U username and -P path to private key"))
				return
			}
			id, err := parseId()
			if err != nil {
				os.Stdout.Write([]byte(err.Error()))
				return
			}
			skShare := mapToKeyShare(id)
			if skShare == nil {
				os.Stdout.Write([]byte("error mapping to G2 curve"))
			}

			g2 := <-lib.EncodeHex(skShare.PublicKeyShare().V.Marshal())
			os.Stdout.Write([]byte("\n" + flagUser + "::" + id.Recipient().String() + "::" + string(g2)))
			os.Stdout.Write([]byte("\n"))
		},
	}

	sendCmd *cobra.Command = &cobra.Command{
		Use:     "send",
		Aliases: []string{"s"},
		Short:   "encrypt and send a message",
		Run:     SendMSG,
	}

	receiveCmd *cobra.Command = &cobra.Command{
		Use:     "receive",
		Aliases: []string{"recv", "r"},
		Short:   "receive and decrypt a message",
		Run:     RecvMSG,
	}
)

func init() {

	// flags for send/receive commands
	sendCmd.PersistentFlags().StringVarP(&flagMsg, "msg", "m", flagMsg, "message to send")
	sendCmd.PersistentFlags().StringVarP(&flagRecip, "recip", "r", flagRecip, "recipient user name")
	sendCmd.PersistentFlags().StringVarP(&flagPath, "path", "p", flagPath, "absolute path to private key")
	receiveCmd.PersistentFlags().StringVarP(&flagMsgId, "msgid", "k", flagMsgId, "message id to receive")
	RootCmd.PersistentFlags().StringVarP(&flagPath, "path", "P", flagPath, "absolute path to private key")

	// Mark flags as required for sending and receiving
	sendCmd.MarkFlagRequired("m")
	sendCmd.MarkFlagRequired("r")
	receiveCmd.MarkFlagRequired("p")
	receiveCmd.MarkFlagRequired("k")

	// Flags for the root cmds
	RootCmd.PersistentFlags().BoolVarP(&flagMinimalOutput, "q", "q", false, "minimal out")
	RootCmd.PersistentFlags().StringVarP(&flagApi, "api", "A", flagApi, "api endpoint")
	RootCmd.PersistentFlags().StringVarP(&flagUser, "user", "U", flagUser, "flag user")
	RootCmd.AddCommand(sendCmd, receiveCmd, infoCmd, mapCmd)
}

// Encrypts and sends a message
func SendMSG(cmd *cobra.Command, args []string) {
	// check that we received the correct flags else return early
	if anyInvalid(flagMsg, flagRecip, flagApi, flagUser) {
		cmd.Help()
		os.Stdout.Write([]byte(ErrIncorrectFlag))
		return
	}

	// TODO: this private key is generated for each message sent and will share
	// its public key with the server despite the server not doing anything with the value
	// at the moment.
	// questions i have;
	//		- should we remove it entirely?
	// 		- maybe we use this as a session key?
	key, _ := age.GenerateX25519Identity()
	c, _ := lib.NewClient(flagApi, flagUser, key)

	// Query for the recipients public key.
	rawPubKey, err := c.Get(flagRecip)
	if err != nil {
		log.Printf("failed to parse public key")
		return
	}

	// Parse the raw public key into an age recipient.
	recips, err := age.ParseRecipients(bytes.NewReader(rawPubKey))
	if err != nil {
		log.Printf("failed to parse recipients")
		return
	}

	if len(recips) < 1 {
		log.Printf("no recipients")
		return
	}

	// Allocate buffer
	encBuffer := bytes.NewBuffer([]byte{})

	// Encrypt to recip public key
	wc, err := age.Encrypt(encBuffer, recips[0])
	if err != nil {
		log.Println("failed to encrypt")
		return
	}

	// Write plaintext message into writercloser buffer.
	if _, err = wc.Write(<-lib.EncodeHex([]byte(flagMsg))); err != nil {
		log.Printf("failed to write buffer")
		return
	}

	// Close writercloser and flush encrypted message to encBuffer
	if err = wc.Close(); err != nil {
		log.Println("error closing buffer")
		return
	}

	id, err := parseId()
	if err != nil {
		panic(err)
	}

	// map the key as point G1 onto pairing curve
	skShare := mapToKeyShare(id)

	// get the public key (G2)
	pubk := skShare.PublicKeyShare().V

	// sign the enc message buffer
	signature := bls.Sign(skShare.V, encBuffer.Bytes())
	hexBuffer := <-lib.EncodeHex(encBuffer.Bytes())

	// verify our signature & message with our public key
	if !bls.Verify(pubk, encBuffer.Bytes(), signature) {
		panic("failed to verify BLS signature")
	}

	// hex encode the signature and use it as the index for the message.
	hexSig := <-lib.EncodeHex(signature.Marshal())

	fmtReq := func(_user, _signature, _mId, _msg string) string {
		return fmt.Sprintf("%s/set?user=%s?sign=%s?mid=%s?msg=%s", flagApi, _user, _signature, _mId, _msg)
	}

	req := fmtReq(flagUser, string(hexSig), string(hexSig[:8]), string(hexBuffer))

	// hex encode the encrypted buffer & set the message key to the resulting
	// value on the server.
	if _, err = http.Get(req); err != nil {
		log.Printf("error sending set request")
		return
	}

	// Write the uuid fo the message to Stdout for receiver
	if !flagMinimalOutput {
		os.Stdout.Write([]byte("\nsent! share this with your recipient:\n\n"))
		os.Stdout.Write(append([]byte("\tlofi r -k "), hexSig...))
		os.Stdout.Write([]byte("\n"))
		return
	}

	os.Stdout.Write([]byte("\n"))
	os.Stdout.Write([]byte(req))
	os.Stdout.Write([]byte("\n"))
}

// Receives and decrypts a message
func RecvMSG(cmd *cobra.Command, args []string) {
	// check that we received the correct flags else return early
	if anyInvalid(flagPath, flagMsgId, flagApi, flagUser) {
		cmd.Help()
		os.Stdout.Write([]byte(ErrIncorrectFlag))
		return
	}
	id, err := parseId()
	if err != nil {
		log.Println(err)
		return
	}
	// setup a new client
	c, err := lib.NewClient(flagApi, flagUser, id)
	if err != nil {
		log.Println("failed to parse client")

		log.Println(err)
		return
	}

	// make https request with the passed in
	// message id https://api.com/get?flagMsgId
	msg, err := c.Get(flagMsgId)
	if err != nil {
		log.Println("Unknown result from flag msg id")
		return
	}

	// hex decode the result
	rd := bytes.NewReader(<-lib.DecodeHex(msg))

	// decrypt with local private key
	rc, err := age.Decrypt(rd, id)
	if err != nil {
		log.Println("failed to decrypt: possibly wrong private key or malformed data")
		return
	}

	buff := bytes.NewBuffer([]byte{})
	// copy decryption to standard output
	_, err = io.Copy(buff, rc)

	// decode hex one last time to get original message
	os.Stdout.Write(<-lib.DecodeHex(buff.Bytes()))

	if err != nil {
		log.Println("failed to copy stdout")
		return
	}

}

func parseId() (*age.X25519Identity, error) {
	// parse the private key of the receiver
	var id *age.X25519Identity
	id, _ = age.GenerateX25519Identity()
	if id == nil {
		return nil, errors.New("failed to parse")
	}

	fd, errno := syscall.Open(flagPath, os.O_RDONLY, 077)
	if errno != nil {
		return nil, errors.New("failed to parse")
	}

	f := os.NewFile(uintptr(fd), flagPath)
	_mat, err := io.ReadAll(f)
	if err != nil {
		log.Printf("parse error: private key at %s", flagPath)
		return nil, errors.New("failed to parse")
	}

	id, err = age.ParseX25519Identity(string(_mat))
	if err != nil {
		log.Println("bad pk parse")
		return nil, errors.New("failed to parse")
	}
	return id, nil
}

func anyInvalid(flags ...string) bool {
	for _, flag := range flags {
		if len(flag) == 0 {
			return true
		}
	}
	return false
}
