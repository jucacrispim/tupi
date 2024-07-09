.. _plugins:

Writing plugins
===============

Tupi supports extention throught plugins using the go's plugin interface. Every plugin
may have an OPTIONAL ``Init`` function for the plugins initialization. The ``Init``
function gets a domain and a reference to a config map and returns an error

.. code-block:: go

   package main

   func Init(domain string, conf *map[string]any) error {
	// do something
	return nil
   }


Serve plugin
------------

To create a serve plugin you must implement a function named ``Serve`` the get
three params: A ``http.ResponseWriter``, a reference to ``http.Request``, a domain and a referece to a
config map.

.. code-block:: go

   package main

   import "net/http"

   func Serve(w http.ResponseWriter, r *http.Request, conf *map[string]any) (bool, int, []byte) {
       w.WriteHeader(200)
       w.Write([]bytes("everything ok!"))
   }


Authentication plugin
---------------------

To create an authentication plugin you must implement a function named ``Authenticate`` that
get three params: A reference to ``http.Request``, a domain and a reference to a
config map and returns a bool and a int indicating if the authentication was successfull
or not and the http status to be returned in case of failed authentication.

.. code-block:: go

   package main

   import "net/http"

   func Authenticate(r *http.Request, domain string, conf *map[string]any) (bool, int) {
	   if r.Host == "test.localhost" {
		   return true, 200
	   }
	   return false, 403
   }


To compile the plugin you need to use ``-buildmode=plugin`` and ``-trimpath`` flags:

.. code-block:: sh

   $ go build -o my_plugin.so -buildmode=plugin -trimpath my_plugin.go


Now in your tupi config file you need to pass the path of the plugin.

.. code-block:: toml

   AuthPlugin = "/path/to/my_plugin.so"
   AuthPluginConf = {
       "something": "the-value"
   }


Check :ref:`config-file` for more information on the tupi config file.


Caveats
-------

Being built upon the go's plugin mechanism, tupi plugins inherit its
caveats:

- Plugins are Linux/BSD only.

- The main application and the plugins must be compiled with the same version
  of the compiler.

- Plugins can't be unloaded and in the case of tupi plugins can't be reloaded.
  Once a plugin is loaded the only way to reload it is restarting the server.
