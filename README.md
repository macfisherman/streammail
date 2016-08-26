# streammail
A new twist on communicating

The idea behind Streammail is that most of the time, communication happens between two parties. Streammail creates a unique address for this communication. Currently this address is derived via the Diffie-Hellman protocol. The address is the resulting *secret* of the protocol exchange. This secret is then converted into a bitcoin like address. This has the advantage of each party knowing the secret yet the server that hosts that address will not. Each party can then encrypt the data send to a Stream storage server. It also allows each party to decrypt messages as well. This means the the smarts of the system are done via clients and the server is as simple storage system with minimal logic.

The overall goal of this project is to create several versions of client/servers using different languages. Besides offering a new communication medium, the project is a way for the author to explore/demonstrate different programming languages. The first implementation will be done in Google's Go.
