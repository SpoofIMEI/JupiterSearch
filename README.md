<img width=500 src="https://github.com/R00tendo/JupiterSearch/assets/72181445/df7259fc-862f-4c47-848a-b53edf473c31"></img>
# JupiterSearch (version 1.3.1)

JupiterSearch is an easy-to-setup distributed text search database that is designed for searching for unique information or keywords like serial numbers, email addresses, and domain names from huge amounts of unstructured data, for example, websites, documents, and emails.

**What JupiterSearch offers you:**
- Easy to set up
- Suitable for unstructured data like emails/documents/web pages
- Can handle terabytes of data
- Client library
- Trivial horizontal scaling

**What JupiterSearch is NOT good for:**
- Relational data
- Extremely sensitive data (because HTTPS is not enabled by default and keys are stored plaintext in conf files)

<br>

**Todo (in chronological order from oldest to newest):**
- [x] Custom tokenization
- [x] HTTPS
- [x] Multiple queries
- [ ] Make repo public 
- [ ] Github wiki

<br>

# Getting started
### Prerequisites
- Go (preferably the latest version)
- Linux -based system
- At least 2GB of disk space

<br>

### Installation
Download JupiterSearch either by using `git clone` or by downloading and unpacking the zip file on this page.

```sh
git clone https://github.com/R00tendo/JupiterSearch
```

<br>

Run `make install` as root to automatically download the dependencies, compile the programs, and install JupiterServer, JupiterNode, and JupiterClient on your system (/usr/local/bin).

```sh
sudo make install
```

<br>

### Config settings (the ones marked with * are required)
- #### Node
  - *`datadir`: Path to where the database will be stored in
  - *`max_concurrent_ingests`: Amount of concurrent store requests that are allowed
  - *`name`: The name that will show as the source for results when you query something
    
- #### Master server
  - *`client_key` <b>(IMPORTANT)</b>: This is essentially the password for the whole system. Clients authenticate using this.
  - *`nodes` <b>(IMPORTANT)</b>: List of nodes separated by a space like this: `nodes=http://127.0.0.1:9192 http://127.0.0.1:9193`

- #### Universal
  - *`api_listen`: What host the rest API will be binded to
  - *`node_key`: A key that the master server will use to authenticate itself to the node
  - `tls_cert`: Location to a public certificate (for encrypted rest API traffic)
  - `tls_private`: Location to a private key (used with tls_cert)

<br>

### Configuring node(s)
Open <b><i>/etc/JupiterSearh/JupiterNode.conf</i></b> with your favorite text editor on the machine you want to use as a node.

When you open the file, you will be greeted with these default settings:
```env
name=main_node
datadir=data
api_listen=127.0.0.1:9192
node_key=JupiterKey
max_concurrent_ingests=5
```

Most of these you can leave to default, but I highly recommend changing the `key`, since if you don't, and bind JupiterNode to all interfaces, anyone on the network could get access to your node. 

#### Making JupiterNode accessible from LAN
Unless you're planning to use JupiterSearch on a single machine that runs both the JupiterServer and JupiterNode, you would want to change `api_listen` to bind all interfaces or just your specific network adapter:
```sh
api_listen=0.0.0.0:9192
```

<br>

### Configuring the master server
Open <b><i>/etc/JupiterSearh/JupiterServer.conf</i></b> with your favorite text editor on the machine you want to use as the master server (the one clients can use to store and query data).

These are the default settings:
```sh
api_listen=127.0.0.1:9190
node_key=JupiterKey
client_key=changeme
nodes=http://127.0.0.1:9192
```

Change the `client_key` to something strong and random. Think of it as an API key. A client that has it can do everything.

If you changed `node_key` from the defaults in the node configs, set the same key as a value for `node_key` on the server configs as well.

Add your nodes to the `nodes` variable, separated by a space character.

<br>

### Customizing the tokenization
By default, JupiterSearch extracts all the words and other information by running this regex against the data: `[\w+.+_+@]{4,}`.

However, you can customize this by editing the regex found in <b><i>/etc/JupiterSearch/tokenization_regex</i></b>.

# Usage
There are two ways you can run JupiterNode and JupiterServer.
- As a service
- Commandline
I recommend first running both on the commandline with the `--debug` flag to make sure everything is working, but after that, it would be easier to run them as a service.
#### Commandline
JupiterServer:
```
JupiterServer --start --debug
```

JupiterNode:
```
JupiterNode --start --debug
```

#### Service
JupiterServer:
```
systemctl start JupiterServer
```

JupiterNode:
```
systemctl start JupiterNode
```

Remember to run JupiterNode first, since JupiterServer tries to connect to all the nodes within the config file, and if it is unsuccessful, it will ignore the node(s).

<br>

### JupiterClient
Unless you want to code a client yourself, using JupiterClient is a solid option for manually operating JupiterSearch.

JupiterClient syntax:
```sh
JupiterClient --server <master server url> --key <client_key> <arguments>
```

Example:
```sh
JupiterClient --server http://127.0.0.1:9190 --key 3ms9dk2lfhs83bf9s20 --upload movies.json
```

<br>

# Understanding JupiterSearch
### JupiterSearch parts
JupiterSearch consists of three parts:
- The client
- The master server
- The node(s)
  
#### Client
By client, I refer to any program that wants to store or query data from JupiterSearch. This could be the official client (JupiterClient) or another one that someone built using the client library.

#### Master server
The master server is the service clients interact with. It keeps track of all the nodes, removes inactive ones, and makes sure that the data is equally spread out among all the nodes.

#### Node(s)
Node is the service that actually has the data and can query it. It receives commands/requests from the master server and responds to them appropriately.


<img width=700 src="https://github.com/R00tendo/JupiterSearch/assets/72181445/04f567f8-b517-49cd-99db-f19fb3bc54ce"></img>

<br>

### Data storage
JupiterSearch uses <a href="https://github.com/dgraph-io/badger">Badger</a> as its underlying database.

When the master server receives a document to be stored, this is what happens in the backend:
1. Master server: Looks at all the node(s) database sizes, picks one with the smallest database, and forwards the request to it.
2. Node: Stores the full document in the database with a unique ID.
3. Node: Converts the document to lowercase, tokenizes it (gets all words from it), and removes duplicates.
4. Node: Loops through all the words/usernames/emails and stores them with the ID of the full document.

<br>

#

### Contributions are welcome 
