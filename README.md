# Raft tick

Every server runs the same time tick, but based on his ID, every server only shows some seconds marks in the log.

Every node contains all the information but the responsibility of show to the log (the execution) is distributed.

When a new server is added the responsibility is redistributed.

Start leader

    go run . -id node0 raft0

Start peers

go run . -id node1 -haddr :11001 -raddr :12001 -join :11000 raft1
go run . -id node2 -haddr :11002 -raddr :12002 -join :11000 raft2


Based on

https://github.com/otoolep/hraftd
