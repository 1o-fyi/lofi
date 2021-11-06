package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"filippo.io/age"
	"git.sr.ht/~lofi/lib"
	"github.com/spf13/cobra"
)

const (
	ErrIncorrectFlag = "setup: incorrect or missing flag(s)"
)

var (
	flagMsg   = ""
	flagRecip = ""
	flagPath  = ""
	flagApi   = "https://1o.fyi"
	flagUser  = "nobody"
	flagMsgId = ""

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
	receiveCmd.PersistentFlags().StringVarP(&flagPath, "path", "p", flagPath, "absolute path to private key")
	receiveCmd.PersistentFlags().StringVarP(&flagMsgId, "msgid", "k", flagMsgId, "message id to receive")

	// Mark flags as required for sending and receiving
	sendCmd.MarkFlagRequired("m")
	sendCmd.MarkFlagRequired("r")
	receiveCmd.MarkFlagRequired("p")
	receiveCmd.MarkFlagRequired("k")

	// Flags for the root cmds
	RootCmd.PersistentFlags().StringVarP(&flagApi, "api", "A", flagApi, "api endpoint")
	RootCmd.PersistentFlags().StringVarP(&flagUser, "user", "U", flagUser, "flag user")
	RootCmd.AddCommand(sendCmd, receiveCmd, infoCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func anyInvalid(flags ...string) bool {
	for _, flag := range flags {
		if len(flag) == 0 {
			return true
		}
	}
	return false
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
	if _, err = wc.Write([]byte(flagMsg)); err != nil {
		log.Printf("failed to write buffer")
		return
	}

	// Close writercloser and flush encrypted message to encBuffer
	if err = wc.Close(); err != nil {
		log.Println("error closing buffer")
		return
	}

	// Grabs 256 bits of entropy
	uuid := <-lib.EncodeHex(lib.Rpb(256))

	// Grabs 4 bytes of the public key, starting at the 4th byte of the key
	strK := fmt.Sprintf("%s", recips[0])[4:8]
	msgK := append([]byte(strK+"-"), uuid[:4]...)

	// hex encode the encrypted buffer & set the message key to the resulting
	// value on the server.
	_, err = c.Set(string(msgK), string(<-lib.EncodeHex(encBuffer.Bytes())))
	if err != nil {
		log.Printf("error sending set request")
		return
	}

	// Write the uuid fo the message to Stdout for receiver
	os.Stdout.Write([]byte("\nsent! to receive your message run:\n\n"))
	os.Stdout.Write(append([]byte("\tlofi r -k "), msgK...))
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

	// parse the private key of the receiver
	var id *age.X25519Identity
	var err error = errors.New("failed to parse private key")
outerloop:
	for _, k := range lib.NewDirectoryGraph(flagPath).OpenAll(lib.EXT_AGE) {
		sc := bufio.NewScanner(k)
		for sc.Scan() {
			id, err = age.ParseX25519Identity(string(sc.Bytes()))
			if err != nil {
				continue
			}
			err = nil
			break outerloop
		}
	}
	if err != nil {
		log.Println("bad pk parse")
		return
	}

	// setup a new client
	c, err := lib.NewClient(flagApi, flagUser, id)
	if err != nil {
		log.Println("failed to parse client")
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

	// copy decryption to standard output
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		log.Println("failed to copy stdout")
		return
	}

}
