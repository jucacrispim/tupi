Tupi - A simple http server
============================
.. Fuck you github!!
.. raw:: html

    <img src=https://raw.githubusercontent.com/jucacrispim/tupi/master/logo.svg"" height="100px">

Tupi is a very simple http server. Its main purpose is to provide an easy
way to serve files from, and upload file to, a directory.


Build & install
---------------

Tupi is written in go so you need the go compiler installed & runtime. With
those installed clone the code:

.. code-block:: sh

   $ git clone https://github.com/jucacrispim/tupi


And install the program with:

.. code-block:: sh

   $ cd tupi
   $ make install


.. code-block:: note

   For all make targets use ``make help``


And now you can start the server using the command ``tupi``

.. code-block:: sh

   $ tupi

This is going to serve the files in the default directory in the port
8080

Use the option ``-h`` for full information

.. code-block:: sh

   $ tupi -h


Uploading files
---------------

To upload files is required an authenticated request using basic http auth.
Tupi reads the user auth information from a htpasswd file. To create a
htpasswd file use:

.. code-block:: sh

   $ htpasswd -c -B /my/htpasswd myusername

And start tupi with the ``-htpasswd`` flag:

.. code-block:: sh

   $ tupi -htpasswd /my/htpasswd


.. warning::

   Your htpasswd file must not be within the root directory being served
   by tupi


Using https
-----------

To use https you need to start tupi with ``-certfile`` and ``-keyfile``
flags.

.. code-block:: sh

  $ tupi -certfile /my/file.pem -keyfile /my/file.key
