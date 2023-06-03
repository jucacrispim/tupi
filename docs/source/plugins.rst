Writing plugins
===============

Tupi supports authentication plugins using the go's plugin interface. To create
an authentication plugin you must implement a function named ``Authenticate`` that
get two params: A reference to ``http.Request`` and a ``map[string]any`` that
contains the specific configs for the plugin. The function must returns a
bool indicating if the request was successfully authenticated or not.

.. code-block:: go

   package main

   import "net/http"

   func Authenticate(r *http.Request, conf map[string]any) bool {
	   if r.Host == "test.localhost" {
		   return true
	   }
	   return false
   }


To compile the plugin you need to use ``-buildmode=plugin`` flag:

.. code-block:: sh

   $ go build -o my_plugin.so -buildmode=plugin my_plugin.go


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
