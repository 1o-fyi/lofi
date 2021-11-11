package cmd

const (
	__MOD   string = "lofi"
	__USAGE string = `
Send a (-m) message to a (-r) recipient 

	$ lofi -U your_username -P /path/to/secret.key s -m "hi john" -r john
	> sent! to receive your message run:
	> 		lofi r -k g5j2-0d

Share the receiving command with your friend,
to decrypt the message they'll need to also pass
in the absolute filepath to their private key.

	$ lofi -U john -P /path/to/secret.key r -k g5j2-0d
	> hi john
The (-r) recipient must have a matching public key registered

see ./key-gen.sh && https://github.com/1o-fyi/register

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
to a username. This is for testing/development/learning purposes only.

see ./key-gen.sh && https://github.com/1o-fyi/register

-		-		-		-		-		-

` + __USAGE
)
