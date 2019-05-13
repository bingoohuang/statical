# statical

[![Build Status](https://travis-ci.org/bingoohuang/statical.svg?branch=master)](https://travis-ci.org/bingoohuang/statical)

statical allows you to embed a directory of static files into your Go binary to be later served from an http.FileSystem.

Is this a crazy idea? No, not necessarily. If you're building a tool that has a Web component, you typically want to serve some images, CSS and JavaScript. You like the comfort of distributing a single binary, so you don't want to mess with deploying them elsewhere. If your static files are not large in size and will be browsed by a few people, statical is a solution you are looking for.

## Usage

Install the command line tool first.

	go get github.com/bingoohuang/statical

statical is a tiny program that reads a directory and generates a source file that contains its contents. The generated source file registers the directory contents to be used by statical file system.

The command below will walk on the public path and generate a package called `statical` under the current working directory.

    $ statical -src=/path/to/your/project/public

In your program, all your need to do is to import the generated package, initialize a new statical file system and serve.

~~~ go
import (
  "github.com/bingoohuang/statical/fs"

  _ "./statical" // TODO: Replace with the absolute import path
)

// ...

  staticalFS, err := fs.New()
  if err != nil {
    log.Fatal(err)
  }

  http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(staticalFS)))
  http.ListenAndServe(":8080", nil)
~~~

Visit http://localhost:8080/public/path/to/file to see your file.

There is also a working example under [example](https://github.com/bingoohuang/statical/tree/master/example) directory, follow the instructions to build and run it.

Note: The idea and the implementation are hijacked from [camlistore](http://camlistore.org/). I decided to decouple it from its codebase due to the fact I'm actively in need of a similar solution for many of my projects.
