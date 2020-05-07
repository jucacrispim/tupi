Tupi - A simple http server
============================

Tupi is a very simple http server. Its main purpose is to provide an easy
way to serve files from a directory.

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
