# How to use repclient

Repclient is trivial to use, as long as you only need it for trivial use-cases.
The standard usage is to post something to a server, or retrieve it.


## Trivial use

Posting works like this:

	repclient -in FILE

This will take a few seconds since repclient will generate a hashcash collision and has to transfer the data to the server.
As a response you will receive output like this:

	Pastebin Address: http://bvuk3xmvslx3idcj.onion/3x77hJtt42MkGbs18e1ZvBw9oAftAUrr9K9x4E8rQzed_2PGBikD5hEcXh7kT4vtKPsZuwymWMeBNeGiRpQ24upB3

Simply give the Pastebin Address to whoever should gain access to the file.

To fetch the file from the server, simply enter

	repclient -out OTHERFILE  http://bvuk3xmvslx3idcj.onion/3x77hJtt42MkGbs18e1ZvBw9oAftAUrr9K9x4E8rQzed_2PGBikD5hEcXh7kT4vtKPsZuwymWMeBNeGiRpQ24upB3

Keep in mind to use your own pastebin address. The example address shown will
not work. You will know have the content of the paste in file `OTHERFILE`.

An even easier way accomplish the same task is to use pipes.
Post:

	cat FILE | repclient

Fetch:

	 repclient http://bvuk3xmvslx3idcj.onion/3x77hJtt42MkGbs18e1ZvBw9oAftAUrr9K9x4E8rQzed_2PGBikD5hEcXh7kT4vtKPsZuwymWMeBNeGiRpQ24upB3

The last command will display the contents of the paste on the standard output.

### Custom paste server

You can use a non-default server by adding the `--server URL` option to the commandline. The URL is an onion URL that needs to point to a repbin server. For example:

	cat FILE | repclient --server http://bvuk3xmvslx3idcj.onion/


## Custom config file

To simplify use, speed up post/fetch, and to allow many servers, create a config-file:

	repclient --peerlist

You can edit `~/.config/repclient/repclient.config` to change the address of
your Tor client, or set up a keydir. You should from time to time call
`repclient --peerlist` again so that your paste-servers stay current.

### Keydir

Most of the time spent when pasting is used on creating hashcash tokens. Change
the "KeyDir" option in your new configuration file to point to a new empty
directory. Now start a token generator to fill it with data:

	reptoken --outDir KeyDir

You computer will now start producing hashcash tokens. Whenever you want to
post, the repclient will try to find a usable token in your KeyDir and use that.
Considerable speedup!


## Advanced/expert usage

While the simple model requires the poster to somehow get the URL to the
fetcher, repbin also supports an email-like messaging model.  To use it, the
fetcher (receiving party) needs to generate a long-term key, give it to
potential posters, and occasionally check the server for new messages.


### Create a key

To generate a new long-term key, execute:

	repclient --genkey

This will generate output, like this:

	PRIVATE key: CoxBwGcVTvzt9iEsDMbmGUxLgWCJeeQo9gUTmjzcLmaM

Never ever share that key with anybody. It needs to be kept secret.

For every person you want to communicate with, create a temporary key:

	repclient --gentemp

This will ask for "Private key(s):". Copy the output from the previous step into
the prompt. Two lines of output will be displayed:

	PRIVATE key: CoxBwGcVTvzt9iEsDMbmGUxLgWCJeeQo9gUTmjzcLmaM_oUZsHqsdGaNTjTxFB3r5J5RXx9MYrjkCsrfd9UT4RuJ

	Public key: 8TwsRs53VgTtLiKKvrD1wT5wdZECjGmV29BUtAQAv7V2_FHFi2PLkHzgCEqTyKxZCZZwTcDr7BMwGkAr4wCUGT7Xp

The public key can be given to the sending party. You will need to keep the
private key for yourself. Never share it. And if you lose it you will lose
access to all messages sent to it.

### Send to a key

To send a paste to somebody who has given you his public key, enter this
command:

	cat FILE | repclient --recipientPubKey 8TwsRs53VgTtLiKKvrD1wT5wdZECjGmV29BUtAQAv7V2_FHFi2PLkHzgCEqTyKxZCZZwTcDr7BMwGkAr4wCUGT7Xp

This will send a message that can only be read by the recipient. Also, you can
assure the other party of your identity by adding your own privatekey:

	cat FILE | repclient --privkey CoxBwGcVTvzt9iEsDMbmGUxLgWCJeeQo9gUTmjzcLmaM --recipientPubKey 8TwsRs53VgTtLiKKvrD1wT5wdZECjGmV29BUtAQAv7V2_FHFi2PLkHzgCEqTyKxZCZZwTcDr7BMwGkAr4wCUGT7Xp

In both cases the returned paste-address will look different than in trivial
usage.


### Check for new messages

Check for messages sent to you:

	repclient --index --server http://bvuk3xmvslx3idcj.onion/

It will require you to enter your private key again. The output will look like
this:

	Index		MessageID
	------------------------------------------------------------
	0		    G9MuGQ1wYCXag52BeRjVjQuZHfbqcoe7o9FDwpTC9YJ9
	------------------------------------------------------------

You can either download all messages by adding the `--outdir DIR` option, or you
can manually fetch them


### Manually fetch a message sent to you

As long as you have the MessageID, just enter:

	repclient MessageID

or specifically

	repclient G9MuGQ1wYCXag52BeRjVjQuZHfbqcoe7o9FDwpTC9YJ9

It will ask you for the LONG temporary private key. After decryption the content
will be shown.

Last but not least...


### Repost
	
There exist repost servers that allow you to anonymize a message by sending the
message to the repost server and that server will repost it to the public.
We will only hint at their usage here:

	repclient --encrypt --repost --appdata -minDelay=1200 -maxDelay=7200 | repclient --encrypt --messageType=3 --recipientPubKey <repost server public key>

Have fun....
