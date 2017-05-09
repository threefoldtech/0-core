## Unix sockets inside a container
We already know how to port forward tcp ports from host to a container, it's simply done via the client. But 
a problem arise when we try to do the same for unix-sockets

Unix sockets are different because u can't do the forwarding without having the a server listening in the first place
which puts us in a chicken/egg problem.

One of the suggested solutions was to:
- Create a container
- Start the service that listens on unix socket, inside the container
- Do forwarding from host to container by another call to do the required binding.

This it sounds very complicated an unpractical to implement, instead we took a different approach.

Since the container processes runs after all in a chroot on the host, we theoretically don't need to do 
any forwarding, since the unix-socket still accessible over VFS.

Example:
 - Start an ubuntu container
 - Start `nc -l -U /tmp/unix.socket` inside the conatiner
 - Using the client, u can tell where the root of the container is (let's say it's under /mnt/container-1/)
 - Then u can connect to that socket over full path `/mnt/container-1/tmp/unix.socket`
  
Without the need to do forwarding.