# distributed_chatroom

To run this chatroom service, at the root directory, run go run cmd/chat-server/main.go [server name] [port number] [total num of people in the chatroom]


To set mode, change DEBUG variable in cmd/chat-server/main.go to be TRUE/FALSE
If in debug mode, then it runs in localhost. The port number should set from 5800, 5900, 6000, ...

If in production mode, then the chatroom can run across different ip address instance. The port num should be 5600 in this case. 
