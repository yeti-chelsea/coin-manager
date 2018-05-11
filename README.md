# coin-manager
Manage coin mining, display information about various hardware and its mining speed.

# Basic Design
There will be a https server and https client

What does Https Server do ?
---------------------------
This Http server has 2 main functinality
1. Recieves hear-beat from all mining hardwares.
2. Recieves incoming connection from Web-ui for displaying all the information.

What kind of information does it recieves from mining harwares ?
----------------------------------------------------------------
When the client intially connects it sends out the following information via udp
1. Hostname
2. Ip-address 
3. Number of Cpus
4. Which coin it is mining 
5. Wallet Id 
6. Pool Info.
7. Link to actual coin.

After initial startup, every 5 seconds it publishes the following info
1. State of miner program
2. Log output of miner program


What kind of information does it recieve from Web-UI ?
------------------------------------------------------
Http server should do basic authentication of username and password, Once authentication is done it should display all info
regarding the mining harware in a tabular format

A link to stop and delete all the mining process.
A link to stop and start a different coin.
A link to the pool.
A link to actual website of the coin.

java-script or node.js would be used for designing.
