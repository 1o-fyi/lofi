package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"filippo.io/age"
	"git.sr.ht/~lofi/lib"
	"github.com/spf13/cobra"
)

var (
	RootCmd *cobra.Command = &cobra.Command{
		Use:   __USAGE,
		Short: __SHORT,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			os.Stdout.Write([]byte("\n"))
		},
	}

	infoCmd *cobra.Command = &cobra.Command{
		Use:   "info",
		Short: "learn about lofi cli",
		Run: func(cmd *cobra.Command, args []string) {
			os.Stdout.Write([]byte(__INFO))
		},
	}

	sendCmd *cobra.Command = &cobra.Command{
		Use:   "s",
		Short: "encrypt and send a message",
		Run:   SendMSG,
	}

	receiveCmd *cobra.Command = &cobra.Command{
		Use:   "r",
		Short: "receive and decrypt a message",
		Run:   RecvMSG,
	}
)

var (
	defaultMsg   = ""
	defaultRecip = ""
	defaultPath  = ""
	defaultApi   = "https://1o.fyi"
	defaultUser  = "nobody"
	defaultMsgId = ""
)

func init() {

	// flags for send/receive commands
	sendCmd.PersistentFlags().StringVarP(&defaultMsg, "msg", "m", defaultMsg, "message")
	sendCmd.PersistentFlags().StringVarP(&defaultRecip, "recip", "r", defaultRecip, "recipients user name")
	receiveCmd.PersistentFlags().StringVarP(&defaultPath, "path", "p", defaultPath, "absolute path to private key")
	receiveCmd.PersistentFlags().StringVarP(&defaultMsgId, "msgid", "k", defaultMsgId, "message id you will receive")

	// Mark flags as required for sending and receiving
	sendCmd.MarkFlagRequired("m")
	sendCmd.MarkFlagRequired("r")
	receiveCmd.MarkFlagRequired("k")
	receiveCmd.MarkFlagRequired("i")

	// Flags for the root cmds
	RootCmd.PersistentFlags().StringVarP(&defaultApi, "api", "A", defaultApi, "api endpoint")
	RootCmd.PersistentFlags().StringVarP(&defaultUser, "user", "u", defaultUser, "default user")
	RootCmd.AddCommand(sendCmd, receiveCmd, infoCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// Encrypts and sends a message
func SendMSG(cmd *cobra.Command, args []string) {
	key, _ := age.GenerateX25519Identity()
	c, _ := lib.NewClient(defaultApi, defaultUser, key)

	rawPubKey, err := c.Get(defaultRecip)
	if err != nil {
		log.Printf("failed to parse public key")
		return
	}

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
	if _, err = wc.Write([]byte(defaultMsg)); err != nil {
		log.Printf("failed to write buffer")
		return
	}

	// Close writercloser and flush encrypted message to encBuffer
	wc.Close()

	// Grabs 4 bits of entropy
	uuid := <-lib.EncodeHex(lib.Rpb(4))

	// Grabs 4 bytes of the public key, starting at the 4th byte of the key
	strK := fmt.Sprintf("%s", recips[0])[4:8]
	msgK := append([]byte(strK+"-"), uuid...)

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

// Decrypts and sends a message
func RecvMSG(cmd *cobra.Command, args []string) {
	// parse the private key of the receiver
	var id *age.X25519Identity
	for _, k := range lib.NewDirectoryGraph(defaultPath).OpenAll(lib.EXT_AGE) {
		raw, err := io.ReadAll(k)
		if err != nil {
			panic(err)
		}
		id, err = age.ParseX25519Identity(string(raw))
		if err != nil {
			panic(err)
		}
	}

	// setup a new client
	c, err := lib.NewClient(defaultApi, defaultUser, id)
	if err != nil {
		panic(err)
	}

	// get the message passed in
	msg, err := c.Get(defaultMsgId)
	if err != nil {
		panic(err)
	}

	// hex decode the result
	rd := bytes.NewReader(<-lib.DecodeHex(msg))

	// decrypt with local private key
	rc, err := age.Decrypt(rd, id)
	if err != nil {
		panic(err)
	}

	// copy decryption to standard output
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		panic(err)
	}

}
