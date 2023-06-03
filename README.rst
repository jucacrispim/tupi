Tupi - A simple http server
============================

.. raw:: html

    <img src="https://raw.githubusercontent.com/jucacrispim/tupi/master/docs/source/_static/logo.svg" height="100px">

Tupi is a very simple http server. Its main purpose is to provide an easy
way to serve files from, and upload file to, a directory.


Build & install
---------------

Tupi is written in go so you need the go compiler installed. With that installed
clone the code:

.. code-block:: sh

   $ git clone https://github.com/jucacrispim/tupi


And install the program with:

.. code-block:: sh

   $ cd tupi
   $ make install


.. code-block:: note

   For all make targets use ``make help``


Usage
-----

Tupi was created to serve and upload files to a directory. So first lets create
a directory with files

.. code-block:: sh

   $ mkdir myfiles
   $ echo "My first file" > myfiles/file.txt


Serving files
+++++++++++++

You can start the server using the command ``tupi``

.. code-block:: sh

   $ tupi -root myfiles

This is going to serve the files in the ``myfiles`` directory and the server
will listen in the port 8080

Use the option ``-h`` for all the options for tupi.

.. code-block:: sh

   $ tupi -h

With the server running we can fetch files from the directory.

.. code-block:: sh

   $ curl http://localhost:8080/file.txt
   My first file

You can also list the contents of a directory:

.. code-block:: sh

   $ curl http://localhost:8080/
   <pre>
   <a href="file.txt">file.txt</a>
   </pre>

You can also, instead of listing the contents of a directory, return the
index.html file in it. To do so use the option ``default-to-index``.

.. code-block:: sh

   $ tupi -default-to-index


Uploading files
+++++++++++++++

To upload files is required an authenticated request using basic http auth.
Tupi reads the user auth information from a htpasswd file. To create a
htpasswd file use:

.. code-block:: sh

   $ htpasswd -c -B /my/htpasswd myusername

And start tupi with the ``-htpasswd`` flag:

.. code-block:: sh

   $ tupi -root myfiles -htpasswd /my/htpasswd


.. warning::

   Your htpasswd file MUST NOT be within the root directory being served
   by tupi

Now you can upload files sending a POST request to the "/u/" path in the server.
The request must have the ``multipart/form-data`` Content-Type header and the
file must be in a input named ``file``.

.. code-block:: sh

   $ curl --user test:123 -F 'file=@/home/juca/powerreplica.jpg' http://localhost:8080/u/
   powerreplica.jpg

   $ curl http://localhost:8080/
   <pre>
   <a href="file.txt">file.txt</a>
   <a href="powerreplica.jpg">powerreplica.jpg</a>
   </pre>


Extracting files
++++++++++++++++

Tupi is capable of extracting ``.tar.gz`` files. To extract files you send a
POST request to the "/e/" path in the server. This request must also have the
``multipart/form-data`` Content-Type header and the file must be in a
input named ``file``.

.. code-block:: sh

   $ curl --user test:123 -F 'file=@/home/juca/test.tar.gz' http://localhost:8080/e/
   bla/
   bla/two.txt
   bla/ble/
   bla/ble/four.txt
   bla/ble/bad.txt
   bla/ble/three.txt
   bla/one.txt

   $ curl http://localhost:8080/
   <pre>
   <a href="bla/">bla/</a>
   <a href="file.txt">file.txt</a>
   <a href="powerreplica.jpg">powerreplica.jpg</a>
   </pre>



Using https
+++++++++++

To use https you need to start tupi with ``-certfile`` and ``-keyfile``
flags.

.. code-block:: sh

  $ tupi -root myfiles -certfile /my/file.pem -keyfile /my/file.key


Config file
++++++++++++

You can use a config file instead of command line options. Check the
documentation `here <https://tupi.poraodojuca.dev/index.html#config-file>`_.


Virtual domains
+++++++++++++++

Tupi also supports the use of virutal domains. Check the virtual domains
documantation `here <https://tupi.poraodojuca.dev/#virtual-domains>`_.

Plugins
-------

Tupi can be exteded by plugins. Check the ``documentation <https://tupi.poraodojuca.dev/plugins.html>`_
on how to write plugins for tupi
