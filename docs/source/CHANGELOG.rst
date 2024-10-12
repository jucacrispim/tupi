Changelog
=========


* v0.13.1

  - Fix http server for redir to https
  - Add WWW-Authenticated response header for bad basic auth

* v0.13.0

  - Add alternative port and redir to https config options

* v0.12.0

  - Refactor serve plugin

* v0.11.0

  - Auth refactor
  - Add serve plugins support

* v0.10.1

  - Fix bad auth status code

* v0.10.0

  - Add support to authenticated downloads

* v0.9.2

  - Refactor on file extraction

* v0.9.1

  - Fix prevent overwrite

* v0.9.0

  - Add prevent overwrite config param
  - Add support for upload prefix

* v0.8.0

  - Add loglevel config param

* v0.7.0

  - Refactor on auth plugins interface. Now Authenticate must return an int to be
    used as an http reponse in case of failed authentication

* v0.6.4

  - Fix make install

* v0.6.3

  - Pre process config file so inline tables can have new lines

* v0.6.2

  - Change cmd path so it can be installed using go install

* v0.6.1

  - Fix load plugins

* v0.6.0

  - Refactor plugins to allow plugin initialization

* v0.5.0

  - Add support for authentication plugins

* v0.4.0

  - Add support for virtual domains
  - Fix default to index

* v0.3.0

  - Add support for toml config file

* v0.2.0

  - Add -default-to-index option
  - Fix deadlock

* v0.1.0

  - First tupi version
