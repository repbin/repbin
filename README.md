# Repbin: Replicated Encrypted Paste Bin

Repbin is an encrypted pastebin for the command line that runs over Tor!
Repbin servers form a distributed network where nodes sync posts with each other
(like in Usenet or BBS/Fido systems). This makes Repbin resilient and scalable.
Repbin focuses on privacy (encrypted messages) and anonymity (padding and
repost chains).

Post a file:

	cat FILE | repclient

As a response you will receive output like this:

	Pastebin Address: http://bvuk3xmvslx3idcj.onion/3x77hJtt42MkGbs18e1ZvBw9oAftAUrr9K9x4E8rQzed_2PGBikD5hEcXh7kT4vtKPsZuwymWMeBNeGiRpQ24upB3

Simply give the Pastebin Address to whoever should gain access to the file.
Fetch:

	repclient http://bvuk3xmvslx3idcj.onion/3x77hJtt42MkGbs18e1ZvBw9oAftAUrr9K9x4E8rQzed_2PGBikD5hEcXh7kT4vtKPsZuwymWMeBNeGiRpQ24upB3


## Installation

Client software to send and receive file:

	go get -u github.com/repbin/repbin/cmd/repclient

Server software that makes posts available to users:

	go get -u github.com/repbin/repbin/cmd/repserver

Tool to generate hashcash tokens that are required for posting to repservers:

	go get -u github.com/repbin/repbin/cmd/reptoken


## Features

- Forward secure encryption of posts using DHE-curve25519. Even a compromised
  long-term key does not allow to decrypt old posts.
- Integrity protection of posts using HMAC-SHA256. You can be sure that posts
  have not been tampered with.
- Confidentiality of posts using AES256-CTR. Without the recipient key, nobody
  can read the post.
- All posts are padded to a common size. That means that posts are not
  distinguishable by their size when looking ``on the wire''.
- Post are replicated between all servers in the Repbin network.
- Optional constant receiver keys for post-box functionality.
- Receiver key attributes for synchronization and post-box authentication.
- Resource control via hashcash (sha256) and ed25519.
- Some privacy protection by using Tor for all communication and ephemeral keys.
- undocumented goodies.


## Post-box functionality

Generate a new long-term key:

	repclient --genkey

This will generate output, like this:

	PRIVATE key: CoxBwGcVTvzt9iEsDMbmGUxLgWCJeeQo9gUTmjzcLmaM

Never ever share that key with anybody. It needs to be kept secret.

For every person you want to communicate with, create a temporary key:

	repclient --gentemp

This will ask for "Private key(s):". Copy the output from the previous step into the prompt. Two lines of output will be displayed:

	PRIVATE key: CoxBwGcVTvzt9iEsDMbmGUxLgWCJeeQo9gUTmjzcLmaM_oUZsHqsdGaNTjTxFB3r5J5RXx9MYrjkCsrfd9UT4RuJ

	Public key: 8TwsRs53VgTtLiKKvrD1wT5wdZECjGmV29BUtAQAv7V2_FHFi2PLkHzgCEqTyKxZCZZwTcDr7BMwGkAr4wCUGT7Xp

The public key can be given to the sending party. You will need to keep the
private key for yourself. Never share it. And if you lose it you will lose
access to all messages sent to it. Update this key frequently (by re-running gentemp) to get the most out of forward secrecy.

The sender uses the public key like this to send messages:

	cat FILE | repclient --recipientPubKey 8TwsRs53VgTtLiKKvrD1wT5wdZECjGmV29BUtAQAv7V2_FHFi2PLkHzgCEqTyKxZCZZwTcDr7BMwGkAr4wCUGT7Xp
	
You can list messages sent to key as follows:

	repclient --index --privkey CoxBwGcVTvzt9iEsDMbmGUxLgWCJeeQo9gUTmjzcLmaM_oUZsHqsdGaNTjTxFB3r5J5RXx9MYrjkCsrfd9UT4RuJ


## Running a server

If you are an experienced UNIX sysadmin, please consider running your own Repbin
server to help the Repbin network.

While running a server requires hardly any interaction, setting up a server in
the Repbin network requires at least one manual peering agreement with another
server in the network. This is a time-tested architecture which is used
successfully to run the Internet, the Usenet, and BBS networks like FidoNet. To
set up a peering you have to exchange public peering keys with another server
and configure your server accordingly.

To get in touch with us for peering send a message to
`7VW3oPLzQc7VS2anLyDtrdARDdSwa7QTF7h3N2t6J2VN_AjWZQfHoqK3yNqvXPkcswLNXSzFrCzJuRRKZKvY71UWT`
and don't forget to put your own key into the message.

The server installation and the peering process is documented in detail here:
[doc/SERVER-INSTALL.md](https://github.com/repbin/repbin/blob/master/doc/SERVER-INSTALL.md)


## Here be dragons...

Dive deeper into the documentation and the code, if you want to figure out how
to send repost messages (remailer style) and how to run your own reposter
service!


## Further documentation

- How to use the client: [USAGE.md](https://github.com/repbin/repbin/blob/master/USAGE.md)
- How to use reptoken: [doc/REPTOKEN.md](https://github.com/repbin/repbin/blob/master/doc/REPTOKEN.md)
- How to compile: [doc/COMPILE.md](https://github.com/repbin/repbin/blob/master/doc/COMPILE.md)
- How to install server: [doc/SERVER-INSTALL.md](https://github.com/repbin/repbin/blob/master/doc/SERVER-INSTALL.md)
- Design details: [doc/DESIGN.md](https://github.com/repbin/repbin/blob/master/doc/DESIGN.md)
- Application integration: [doc/INTEGRATION.md](https://github.com/repbin/repbin/blob/master/doc/INTEGRATION.md)
- Contributed tools: [contrib/TOOLS.md](https://github.com/repbin/repbin/blob/master/contrib/TOOLS.md)


## Requirements

Client:
- Running Tor client
- Unix-ish operating system (tested on Linux Debian, Gentoo, Ubuntu, recent versions)
- Other operating systems should work except for OS-dependent features like tty support
- Compilation: go >= 1.4

Server:
- Running Tor client and ability to configure a hidden service
- Unix operating system
- Filesystem supporting chtime and symbolic links
- A lot of storage space
- Constant internet connection
- Sysadmin know-how. Really
- Compilation: go >= 1.4


## WARNING

THIS SOFTWARE HAS NEVER BEEN AUDITED OR REVIEWED. IT HAS NOT BEEN TESTED. THE
AUTHORS ARE AMATEURS AND YOU SHOULD NOT USE THIS SOFTWARE FOR ANYTHING
IMPORTANT. YOU SHOULD NOT RELY ON THE SOFTWARE TO WORK AT ALL, OR IN ANY
PREDICTABLE WAY, NOR SHOULD YOU ASSUME THAT THE FEATURES CLAIMED ARE THE
FEATURES IMPLEMENTED. THIS SOFTWARE IS FULL OF ERRORS, THE ARCHITECTURE AND
DESIGN ARE BROKEN. UNLESS SOME EXPERT CLAIMS OTHERWISE.
