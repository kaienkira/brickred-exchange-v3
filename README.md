brickred-exchange-v3
====================

Brickred Studio data serialization tool V3

Build Compiler
--------------
Compiler is written in go
```
cd compiler
bash build_linux.sh
bash build_windows.sh
```

Compiler Usage
--------------
```
usage: brexc -f <protocol_file> -l <language> 
    [-o <output_dir>]
    [-I <search_path>]
    [-n <new_line_type>] (unix|dos) default is unix
language supported: cpp php csharp
```

Use with C++
------------
* build c++ brickred exchange library
```
cd cpp
./config.sh --prefix=<prefix>
make release && make install
```

* generate c++ source and header
```
$ brexc -f attr.xml -l cpp
$ brexc -f message_test.xml -l cpp
$ brexc -f message_type.xml -l cpp
```

* we will get c++ source and header files in current dir
```
$ ls -1 *.cc *.h
attr.cc
attr.h
message_test.cc
message_test.h
message_type.cc
message_type.h
```

* write a main.cc to use the generated code (in example/main.cc)
* compile, link and test
```
$ g++ -c attr.cc
$ g++ -c message_test.cc
$ g++ -c message_type.cc
$ g++ main.cc attr.o message_test.o message_type.o -lbrickredexchange
$ ./a.out
```

Use with PHP
------------
* put php/BrickredExchange.php to your project dir

* generate php file
```
$ brexc -f attr.xml -l php
$ brexc -f message_test.xml -l php
$ brexc -f message_type.xml -l php
```

* we will get generated php files in current dir
```
$ ls -l *.php
attr.php
BrickredExchange.php
message_test.php
message_type.php
```

* write a main.php to use the generated code (in example/main.php)
* run main.php
```
$ php main.php
```

Use with C#
---------------
* build csharp brickred exchange library(example use mono)
```
cd csharp
make
```

* generate csharp source
```
$ brexc -f attr.xml -l csharp
$ brexc -f message_test.xml -l csharp
$ brexc -f message_type.xml -l csharp
```

* we will get generated csharp files in current dir
```
$ ls -l *.cs
attr.cs
message_test.cs
message_type.cs
```

* write a main.cs to use the generated code (in example/main.cs)
* compile, link and test
```
mcs main.cs attr.cs message_test.cs message_type.cs -r:Brickred.Exchange.dll
mono main.exe
```
