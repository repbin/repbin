Repbin: Replicated Encrypted Paste Bin
--------------------------------------

Repbin is a set of tools that allow the quick sharing of small files from the commandline.
It consists of libraries and three programs:
- repclient: Client software to send and receive file
- repserver: Server software that makes posts available to users
- reptoken: Tool to generate hashcash tokens that are required for posting to repservers

Features:
- Ease of use for default use-case: Post a file or fetch a post
- Forward secure encryption of posts using DHE-curve25519 
- Integrity protection of posts using HMAC-SHA256
- Confidentiality of posts using AES256-CTR
- Unified file size by padding all posts to a common size
- Replication of posts between connected servers
- Optional constant receiver keys for post-box functionality
- Receiver key attributes for synchronization and post-box authentication
- Resource control via hashcash (sha256) and ed25519
- Some privacy protection by using Tor for all communication and ephemeral keys
- undocumented goodies

Optional features defined by repserver operators:
- One-Time posts that are deleted as soon as they have been fetched by a client
- Ability to immediately delete a post from the server if private keys are known

Requirements, client:
- Running Tor client
- Unix-ish operating system (tested on Linux Debian, Gentoo, Ubuntu, recent versions)
- Other operating systems should work except for OS-dependent features like tty support
- Compilation: go >= 1.4

Requirements, server:
- Running Tor client and ability to configure a hidden service
- Unix operating system
- Filesystem supporting chtime and symbolic links
- A lot of storage space
- Constant internet connection
- Sysadmin know-how. Really
- Compilation: go >= 1.4

Further information:
How to use the client:     USAGE.txt
How to use reptoken:       doc/REPTOKEN.txt
How to compile:            doc/COMPILE.txt
How to install the server: doc/SERVER-INSTALL.txt
Design details:            doc/DESIGN.txt
Application integration:   doc/INTEGRATION.TXT
Contributed tools:         contrib/TOOLS.txt

WARNING: THIS SOFTWARE HAS NEVER BEEN AUDITED OR REVIEWED. IT HAS NOT BEEN TESTED.
         THE AUTHORS ARE AMATEURS AND YOU SHOULD NOT USE THIS SOFTWARE FOR ANYTHING IMPORTANT.
         YOU SHOULD NOT RELY ON THE SOFTWARE TO WORK AT ALL, OR IN ANY PREDICTABLE WAY, NOR
         SHOULD YOU ASSUME THAT THE FEATURES CLAIMED ARE THE FEATURES IMPLEMENTED.
         THIS SOFTWARE IS FULL OF ERRORS, THE ARCHITECTURE AND DESIGN ARE BROKEN.
         UNLESS SOME EXPERT CLAIMS OTHERWISE.

Note:
SIGNATURE.asc was produced by running:
find * -type f -follow -exec sha256sum \{\} \; | sort | gpg --armor --detach-sign -u 0xDCA2BA86 > SIGNATURE.asc
