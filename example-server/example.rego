package apis

import input.method
import input.authToken

default invocation_allowed = false

invocation_allowed = true {
    internalClient
}

invocation_allowed = true {
    externalClient
    commands[_] = input.method
}

commands = [
   "/helloworld.Greeter/SayHello"
]

clients = {
    "123": {"external": false},
    "234": {"external": false},
    "456": {"external": true},
    "567": {"external": true},
}

internalClient {
    not clients[input.authToken].external
}

externalClient {
    clients[input.authToken].external
}

