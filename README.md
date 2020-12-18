# mp707

mp707 is a library (and a tool) for interacting with MP707 USB temperature sensor boards.

## Build and install

### mp707 utility

Building requires [libusb](https://libusb.info) (~1.0) shared library and development files to be installed in your system.

Corresponding packages for common Linux distributions are:
* Ubuntu 12.04, 16.04, 18.04, 20.04: `libusb-1.0-0`, `libusb-1.0-0-dev`
* Debian 8, 9, 10: `libusb-1.0-0`, `libusb-1.0-0-dev`
* CentOS 7, 8: `libusbx`, `libusbx-devel`
* Alpine Edge: `libusb`, `libusb-dev`

Run `make` to compile `mp707` utility (a binary would be located in `cmd/mp707` directory).

Run `make static` to compile a statically linked `mp707` binary. 
Building static binaries may require [udev](https://www.freedesktop.org/software/systemd/man/udev.html)
(or [eudev](https://wiki.gentoo.org/wiki/Project:Eudev)) development files to be present in the system.

Copy the compiled binary in one of directories listed in your `$PATH`.

## Usage

### mp707 utility

`mp707` is intended to be used for reading temperature sensors readings.
```
$ mp707 -h
Usage of mp707:
  -device uint
        device ID
  -help
        show help
  -sensor string
        sensors ROM ID
$ 
```

When launching without arguments, the tool scans all available MP707 devices and prints all sensors readings:
```
$ mp707 
2020/12/18 19:14:06 processing device 3
2020/12/18 19:14:09 processing sensor ea011933877d9828
ea011933877d9828 7.06 °C
2020/12/18 19:14:09 processing sensor 7a00000c90526528
7a00000c90526528 18.81 °C
2020/12/18 19:14:09 processing sensor 1800000c911b9228
1800000c911b9228 18.75 °C
$ 
```

One can read specific sensor reading by using corresponding command line flag:
```
$ mp707 -sensor ea011933877d9828
2020/12/18 19:14:12 processing device 3
2020/12/18 19:14:12 processing sensor ea011933877d9828
ea011933877d9828 7.06 °C
$ 
```

## MP707 device history

In the past years the device has been sold under different names:

* BM1707 by [Serg Home Studio](http://usbsergdev.narod.ru/BM1707/BM1707.html) (c. 2009 - c. 2010)
* MP707 by [Masterkit](https://web.archive.org/web/20100729172517/http://www.masterkit.ru/main/set.php?code_id=565375),
           [Olimp](https://olimp-z.ru/mp707) (c. 2010 - c. 2016)
* [Rodos 5S](https://silines.ru/rodos-5s), [Rodos 5Z](https://silines.ru/rodos-5z) by Silines (c. 2017 - present)

Despite these changes, it seems that circuit and firmware remained nearly the same across all device modifications,
and the only thing that varies is VID/PID.

## Remarks

### udev rules to access the device for unprivileged users

Create `/etc/udev/rules.d/80-mp707.rules` file with the following contents:

```
SUBSYSTEM=="usb", ATTRS{idVendor}=="20a0", ATTRS{idProduct}=="4173", GROUP="plugdev", TAG+="uaccess"
```

This should allow members of `plugdev` group access the device without using `sudo`.

## Links

* [DS1825 Programmable Resolution 1-WireDigital Thermometer With 4-Bit ID (datasheet)](https://datasheets.maximintegrated.com/en/ds/DS1825.pdf)