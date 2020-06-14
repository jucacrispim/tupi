Tupi http server
================

![tupi](https://raw.githubusercontent.com/jucacrispim/tupi/master/logo.svg)

Tupi is a very simple http server. Its main purpose is to provide an easy
way to serve files from, and upload file to, a directory.


Build & install
---------------

Tupi is written in go so you need the go compiler & runtime installed. With
those installed clone the code:

```sh
$ git clone https://github.com/jucacrispim/tupi
```


And install the program with:

```sh
$ cd tupi
$ make install
```

For all make targets use ``make help``


And now you can start the server using the command ``tupi``

```sh
$ tupi
Tupi is serving at 0.0.0.0:8080
```

This is going to serve the files in the default directory in the port
8080

Use the option ``-h`` for full information

```sh
$ tupi -h

Usage of tupi:
  -certfile string
        Path for the tls certificate file
  -daemon
        Runs the server in background
  -epath string
        Path to extract files (default "/e/")
  -host string
        host:port to listen. (default "0.0.0.0:8080")
  -htpasswd string
        Full path for a htpasswd file used for authentication
  -keyfile string
        Path for the tls key file
  -logfile string
        Log file used when running in backgroud (default "tupi.log")
  -maxupload int
        Max size for uploaded files (default 10485760)
  -pidfile string
        Pid file for the background server (default "tupi.pid")
  -root string
        The directory to serve files from (default ".")
  -timeout int
        Timeout in seconds for read/write (default 240)
  -upath string
        Path to upload files (default "/u/")
```


Uploading files
---------------

To upload files is required an authenticated request using basic http auth.
Tupi reads the user auth information from a htpasswd file. To create a
htpasswd file use:

```sh
$ htpasswd -c -B /my/htpasswd myusername
```

And start tupi with the ``-htpasswd`` flag:

```sh
$ tupi -htpasswd /my/htpasswd
```


**Warning**: Your htpasswd file must not be within the root directory
being served by tupi


Using https
-----------

To use https you need to start tupi with ``-certfile`` and ``-keyfile``
flags.

```sh
$ tupi -certfile /my/file.pem -keyfile /my/file.key
```
