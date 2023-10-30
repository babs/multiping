# multiping

`multiping` is a cli tool to monitor several targets at once using pings or tcp probing with optional logging of transition between states in a file.

## Demo

Example with an `unstable-server` that flaps every 6s and a few other targets:

`multiping localhost unstable-server tcp://google.com:80 tcp://[::1]:22`

![Demo01](https://raw.githubusercontent.com/babs/multiping/master/_demo/demo-001.svg)

If you use a unix like shell, you can take advantage of shell's expansion for ranges like so:

`multiping 192.168.0.{1..10}`

## Documentation

See `multiping -h` for detailed information.

Available probing means are:
- pure go ping (pro-bing, default)
- OS's ping command, via background process
- tcp (partial (S/SA/R tcp-shaker) or full handshake depending on the os)

### ping

Pure go is the default option but for unprivileged users ([see linux notes](#linux-notes-on-pure-go-ping)), OS/system's ping command (usually available on OS with specific cap or setuid) can be used with a background spawn model with `-s` flag. Privileged mode (default when user is root or on windows) can be forcefully enabled with `-privileged`.

### TCP probing

For tcp probing, on linux, freebsd and openbsd, S/SA/R pattern is used. This allows to probe tcp ports without really triggering an accept on the listening app. Issue is if a device in between perform syn poxing, the result might not reflect reality.
On darwin and windows due to limitations, complete handshake is performed.

tcp probing example syntax:
- `tcp://google.com:80`
- `tcp://192.168.0.1:443`
- `tcp://[::1]:22`

### Transition logging

Transition logging can be enabled using `-log filename`.
Log format is pretty self explanatory:

* Timestamp (string): timestamp
* UnixNano (int64): timestamp in nano seconds
* Host (string): the host provided as arg (inc. proto)
* Ip (string): the resolved host
* State (bool): true if alive, false if timeout
* Transition (string): "down to up" or "up to down"

### Quiet mode

`-q` disable the refreshing output, might be useful in conjunction with `-log`.

## Linux notes on pure go ping

If run unprivileged, you might need to allow groups to perform "unprivileged" ping via UDP with the following sysctl:
```bash
sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

You can also add net raw cap to the binary to use it with `-privileged` mode
```bash
cap_net_raw=+ep /path/to/your/compiled/binary
```

## Source

Github repository: https://github.com/babs/multiping

### libs used

* https://github.com/pterm/pterm 
* https://github.com/prometheus-community/pro-bing
* https://github.com/tevino/tcp-shaker
