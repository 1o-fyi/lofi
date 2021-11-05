package cmd

const (
	__MOD   string = "lofi-cli"
	__USAGE string = `
Send a (-m) message to a (-r) recipient 

	$ lofi s -m "hi john" -r john
	> sent! to receive your message run:
	> lofi r -k g5j2-0d

Share the receiving command with your friend,
to decrypt the message they'll need to also pass
in the absolute filepath to their private key.

	$ lofi r -k g5j2-0d -p /path/to/my/private_key

The (-r) recipient must have a matching public key registered
to participate. 

Recursers are welcome to register a key, however, please know
that this is for testing & learning purposes only. Any misuse
of 1o.fyi will result in immediate key-revocation. 

Reach out directly to John S for your key!
	`
	__SHORT string = `
       _               
|  _ _|_ o     _  |  o 
| (_) |  |    (_  |  | 

`
	__INFO string = `
Hiya friend, 

This lofi cli was made as a learning experience during 
a batch at the recurse center. 

To participate you'll need to have a public key registered
to a username.

Recursers are welcome to register a key, however, please know
that this is for testing & learning purposes only. I trust 
everybody will be respectful & conscious.

lofi is using age keys, to register a public key 
reach out directly to John S over Zulip.

-		-		-		-		-		-

Send a (-m) message to a (-r) recipient 

	$ lofi s -m "hi john" -r john
	> sent! to receive your message run:
	> 	lofi r -k g5j2-0d

Share the receiving command with your friend,
to decrypt the message they'll need to also pass
in the absolute filepath to their private key.

	$ lofi r -k g5j2-0d -p /path/to/my/private_key

	       _               
	|  _ _|_ o     _  |  o 
	| (_) |  |    (_  |  | 
	
`
)
