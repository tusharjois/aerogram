# aerogram

Simple 1:1 filesharing over a local network using mDNS, with some extra goodies.

## Usage

On sender machine, run

```sh
aerogram < infile.txt
```

or, using the filename as an argument, run

```sh
aerogram --sendfile "infile.txt"
```

On the receiver machine, run

```sh
aerogram > outfile.txt
```

or (just like above) run

```sh
aerogram --recvfile "outfile.txt"
```

and the file should arrive momentarily.

There is also a `--gzip` flag that compresses the file being sent and
decompresses when it is received, in order to reduce network bandwidth.

Because of how aerogram pipes to standard I/O, you can do fun things like pipe
video streams from one machine and play them back on another machine, and let
aerogram handle the creation of the pipe and compression over the network.

## Future Roadmap

This work was largely inspired by
[airpaste](https://github.com/mafintosh/airpaste). However, there are some
additional features I would like to implement:

- [ ] More robust network communication, verified using tests from `testing`.
- [ ] A `--secure` flag that authenticates the other party and encrypts the
      file over the network.
- [ ] A progress indicator while the transfer is occurring.
- [ ] A clear API so other Go clients can send/receive aerograms.

