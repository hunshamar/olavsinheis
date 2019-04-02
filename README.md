Elevator Project
================


Summary
-------
Software for controlling `n` elevators working in parallel across `m` floors. The number of floors, network ports and other relevant constants can be configured in the `config` file. 

Running the program
-------------------
Make sure that either `ElevatorServer` or `SimElevatorServer` is running on a different terminal, as this is not included in the `RunAll`. Then open a terminal in the project folder, and type in the following commands:
- Giving RunAll access: 		`chmod +x *`
- Giving the other files access: 	`chmod +x **/*` 
- Finally running all file: 	`./RunAll`

Layout description
-----------------

### Single Elevator
 - contains finite state machine of a single elevator with corresponding message handler. The message handler handles state updates from its elevator and forwards it to the master if a connection to the master is established. Note that cab orders are handled internally in the finite state machine, and are therefore not passed to the message handler.

### Backup
 - Implementation of the systems master/backup and interaction. During the initialization of an elevator, a backup is spawned. If there are two or more backups on the network, a master is chosen. This master runs a cost-function every time it gets a state update from one of the elevators on the network. It then sends a master message containing a shared lightmatrix and updated order matrices to every elevator message handler on the network.
 
### Elevator Driver
 - Contains structs, types and functions that are used by the finite state machine and the message module (found under `network`). 
 
### Network
 - contains functions and message formats used to establish UDP-communication and message-passing between the message handler of each elevator and the master/backup.
 
### Config
 - For configuration of system constants, like the number of floors and open door time.




Borrowed code
---------------------
The following subfolders of `network` are borrowed from the [Go network module](https://github.com/TTK4145/Network-go):
`bcast`, `conn`, `localip` and `peers`, though slight modifications has been made. Notably the ablility to close a peer transmisson using a channel.
In addition, the I/O-interface along with some of the types in `elevStruct` are borrowed from the handed out [Go-driver](https://github.com/TTK4145/driver-go). It is also worth mentioning that the finite state machine is based on Anders' FSM presented in one of his lab lectures. 
