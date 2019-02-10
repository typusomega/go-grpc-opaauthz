package simple

import input.method
import input.authToken

default invocation_allowed = false

invocation_allowed = true {
    clients[input.authToken][_] = input.method
}

clients = {
    "123": ["/helloworld.Greeter/SayHello"],
    "234": ["/helloworld.Greeter/SayHello", "/admins/Purge"],
    "456": [],
}

