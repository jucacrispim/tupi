Tupi: A simple web server
=========================

Tupi's main purpose is to provide an easy way to upload and download files. It supports
basic http auth for file uploads.


Install
-------

Tupi is written in go, so you need a go compiler installed. Then, clone the code:

.. code-block:: sh

   $ git clone https://github.com/jucacrispim/tupi


And install the program with:

.. code-block:: sh

   $ cd tupi
   $ make install


.. note::

   For all make targets use ``make help``


Usage
-----

Serving files
+++++++++++++

Now you can use the ``tupi`` command to start the server. By default it serves
the current working dir. It can be changed using the param ``-root``:

.. code-block:: sh

   $ tupi -root /some/dir

By default tupi listens in the port 8080. You can change using the param ``-port``:

.. code-block:: sh

   $ tupi -port 8000

Now you can download files

.. code-block:: sh

   $ curl http://localhost:8080/some-file.txt


You can also list the contents of a directory:

.. code-block:: sh

   $ curl http://localhost:8080/


One can instead of listing the contents of a directory, return the
index.html file in it. To do so use the option ``default-to-index``.

.. code-block:: sh

   $ tupi -default-to-index


Uploading files
+++++++++++++++

To upload files a htpasswd file must be created first for authenticated access:

.. code-block:: sh

   $ htpasswd -c -B /some/htpasswd/file some-user


And start the server using the ``htpasswd`` param:

.. code-block:: sh

   $ tupi -htpasswd -htpasswd /some/htpasswd/file

.. danger::

   Your htpasswd MUST NOT be whithin the directory being served by tupi


To upload a file send a POST request to the "/u/" path in the server.
The request must have the ``multipart/form-data`` Content-Type header and the
file must be in a input named ``file``.

.. note::

   The upload path can be changed by a config param.

.. code-block:: sh

   $ curl --user test:123 -F 'file=@/home/juca/powerreplica.jpg' http://localhost:8080/u/


A ``prefix`` can be passed in the request so the file will be saved inside the ``prefix``
directory

.. code-block:: sh

   $ curl --user test:123 -F 'file=@/home/juca/powerreplica.jpg' http://localhost:8080/u/ -F 'prefix=something'



Upload and extract
++++++++++++++++++

Tupi can extract the contets of uploaded tar.gz files. The contents will be
extracted in the root directory being served and the directory structure
in the tar file will be preserved.

To upload and extract the contents o a file send a POST request to the
"/e/" path in the server. The request must also have the
``multipart/form-data`` Content-Type header and the file must be in
a input named ``file``.

.. note::

   The extract path can be changed by a config param.


.. code-block:: sh

   $ curl --user test:123 -F 'file=@/home/juca/package.tar.gz' http://localhost:8080/e/



HTTPS connections
+++++++++++++++++

To use an HTTPS connection, one must use the ``-certfile`` and ``-keyfile`` params:


.. code-block:: sh

   $ tupi -certfile /some/cert.pem -keyfile /some/file.key


For all options available for tupi use the command ``tupi -h``


.. code-block:: sh

   $ tupi -h
   Usage of tupi:
     -certfile string
	   Path for the tls certificate file
     -conf string
	   Path for the configuration file
     -default-to-index
	   Returns the index.html instead of listing a directory
     -epath string
	   Path to extract files (default "/e/")
     -host string
	   host to listen. (default "0.0.0.0")
     -htpasswd string
	   Full path for a htpasswd file used for authentication
     -keyfile string
	   Path for the tls key file
     -loglevel string
        Log level (default "info")
     -maxupload int
	   Max size for uploaded files (default 10485760)
     -port int
	   port to listen. (default 8080)
     -prevent-overwrite
        Prevents over writing existent files
     -root string
	   The directory to serve files from (default ".")
     -timeout int
	   Timeout in seconds for read/write (default 240)
     -upath string
	   Path to upload files (default "/u/")


.. _config-file:

Config file
+++++++++++

Instead of command line options one can also use a ``toml`` configuration file by
using the ``-config`` command line option or by setting the ``TUPI_CONFIG_FILE``
environment variable.

Here is an example of a config file:

.. code-block:: toml

    # all parameters here are optional
    host = "0.0.0.0"
    port = 1234
    rootDir = "/some/dir"
    # timeout in seconds
    timeout = 500
    htpasswdFile = "/some/htpasswd"
    uploadPath = "/u/"
    extractPath = "/e/"
    # defaults to 10 MB
    maxUploadSize = 10485760
    certFilePath = "/some/cert.pem"
    keyFilePath = "/some/file.key"
    defaultToIndex = true

.. _virtual-domains:

Virtual Domains
+++++++++++++++

To use virutal domains one need to configure the domains in the config file. Each
different section of the config file corresponds to a virtual domain handled by tupi.

.. code-block:: toml

   [default]
   port 8080

   [adomain.net]
   rootDir "/some/dir"
   # timeout in seconds
   timeout = 500
   htpasswdFile = "/some/htpasswd"
   uploadPath = "/u/"
   extractPath = "/e/"
   # defaults to 10 MB
   maxUploadSize = 10485760
   certFilePath = "/some/cert.pem"
   keyFilePath = "/some/file.key"
   defaultToIndex = true
   logLevel = "debug"


   [otherdomain.net]
   rootDir "/other/dir"
   # timeout in seconds
   timeout = 500
   htpasswdFile = "/other/htpasswd"
   uploadPath = "/u/"
   extractPath = "/e/"
   # defaults to 10 MB
   maxUploadSize = 10485760
   certFilePath = "/other/cert.pem"
   keyFilePath = "/other/file.key"
   defaultToIndex = false


All options available ara supported by the virtual domains, except ``host``, ``port``
and ``loglevel`` that are only available for the default server.


Plugins
-------

Tupi can be extended by plugins. Check the docs on how to write your own
plugins

.. toctree::
   :maxdepth: 1

   plugins

.. toctree::
   :maxdepth: 1

   CHANGELOG
