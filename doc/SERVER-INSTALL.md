## Installing a repbin server

If you do not know how to handle a unix sytem, stop here. You'll just hurt
yourself in the process.  We assume that you have Tor running and know how to
configure it to run a hidden service. Furthermore you need to be able to create
a new user on the system.

## Basic installation of a standalone server

0. The server's filesystem must support symlinks and chtimes

1. Install/Compile the binary, as described in COMPILE.txt

2. Configure Tor to point a hidden service to 127.0.0.1:Port (Port is the port
   number of the repserver). Make sure that you restart Tor to have it generate
   the hostname of the hidden service.

3. Create a new user specifically and exclusively for repbin usage. It should have minimum access
    to filesystem resources.

4. Change to the new user

5. Create a new directory for the user to store message data in:

```
	STOREDIR=/path/to/userhome/store/
	mkdir $STOREDIR
```

6. Create initial configuration file:

```
	repserver --showconfig > repserver.config
```

7. Edit the configuration file to change the storage path. Change the value of
	 "StoragePath" to the absolute path of $STOREDIR. Make sure the value ends
	 with a slash (/).

8. If needed, change the SocksProxy entry to the URL of your Tor client.

9. Change the values of EnableDeleteHandler and EnableOneTimeHander to true IF
	 you want these features. EnableOneTimeHander adds the ability to delete
	 messages from the server by anybody knowing the Message-ID and the constant
	 private key of the message. This is not a good idea unless you have pressing
	 need to enable this feature. EnableOneTimeHander allows the use of the
	 repserver for storing messages that are deleted as soon as they are fetched
	 the first time. This is of dubious security. We recommend keeping both
	 settings switched off ("false").

10. Change the ListenPort entry to an unused TCP port on localhost. This must be
    the port that the hiddenservice configuration of Tor points to.

11. Change the URL setting to the _full URL_ of your Tor hidden service (it
		should look similar to this: http://abcdefghijkl.onion/" - note the leading
		schema and the trailing slash).

12. Do not modify other entries unless you know exactly what you do. Save your
    changes to file.

13. Start the server:

```
	repserver --configfile repserver.config --verbose --start
```

14. You're done. The server does not go into the background itself.
    Put it into background manually or run it in a screen/tmux session.

Client's need to use the onion address of your hidden service by either changing
their config-files or specifying the --server commandline option.

Please be aware that there is no locking utilised in repserver. Only one
instance should be running at a time. Running multiple instances accessing the
same $STORAGEDIR WILL lead to catastrophic results, including security
nightmares and dead cats.


## Peering with other servers

To replicate posts from/to other servers, peering needs to be configured. This
is a manual process and there will not be a peer discovery feature for repbin.

After starting the repserver, $STOREDIR will contain a new file:
$STOREDIR/peers.config. This file contains the information required to peer
with other servers. To add a peer, create a new JSON list entry (or change the
example) so that:

- PubKey contains the peer's PeeringPublicKey
- URL contains the peer's hidden service address as a full URI

You can change the peers.config file while the server is running. It will be
reloaded automatically. Errors in the format of the file will lead to peering
becoming unavailable until a well-formatted file is reloaded again.


## Administration tasks

repserver should clean up house itself, but it will leave some files forever.
Please plan for enough space. Do NOT move the files or you will destroy the
chtime of the files (unless you know how to use tar, touch, cp etc). The chtime
is required for part of the housecleaning process. Changing it will lead to
files that linger for too long.

After repserver has received his first files and/or peered with other
repservers, the $STOREDIR will contain a number of directories containing
various files. These files are paged textfiles (every entry is of a defined
length). Changing any of the files manually can lead to strange results,
including making the system unusable or breaking limits and access limitations.
Great care is required if manual changes are required.

File that MIGHT required changes are in the peers directory. Every peer has his
own file here that keeps track of peer status. If a peer is reset (by deleting
his posts), the peer status needs to be reset as well. The easiest method is to
find the peer and delete his status file. Other changes are simply dangerous.


## Security

The software is likely insecure in some bad ways, including but not limited to
giving access to files for reading, writing and changing.  It is highly
recommended to run it in a dedicated chroot, a dedicated virtual server, or on a
dedicated machine. You have been warned.


## Hosting a website

The index of the repserver shows a "404 Not found" error by default. If you want
to spice-up your repserver you may create a directory $STOREDIR/static and put
files into it. These files will be served as with any other standard websert. It
does however not support dynamic content, redirects and similar features. Just
create $STOREDIR/static/index.html to speak to the world!

