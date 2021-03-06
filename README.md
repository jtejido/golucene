golucene
========

A [Go](http://golang.org) port of [Apache Lucene](http://lucene.apache.org).

Continuing where balzaczyy left off and implement all of v4.10.4 with additional experimentation on different ranking models.

This is primarily for my personal use case only, DO NOT USE IN PRODUCTION (I just needed the search capabilities of Lucene in native Go and some alternatives found elsewhere just doesn't cut it for me). 


**TO-DO:**
- Query Expansion.
- Finish some more unimplemented bits found here and there.


Why do we need yet another port of Lucene?
------------------------------------------

Since Lucene Java is already optimized to teeth (and yes, I know it very much), potential performance gain should not be expected from its Go port. Quote from Lucy's FAQ:

>Is Lucy faster than Lucene? It's written in C, after all.

>That depends. As of this writing, Lucy launches faster than Lucene thanks to tighter integration with the system IO cache, but Lucene is faster in terms of raw indexing and search throughput once it gets going. These differences reflect the distinct priorities of the most active developers within the Lucy and Lucene communities more than anything else.

It also applies to GoLucene. But some benefits can still be expected:
- quick start speed;
- able to be embedded in Go app;
- goroutine which I think can be faster in certain case;
- ready-to-use byte, array utilities which can reduce the code size, and lead to easy maintenance.

Though it started as a pet project, I've been pretty serious about this.

Dependencies
------------
Go 1.2+

Installation
------------

	go get -u github.com/jtejido/golucene

Usage
-----

A detailed example can be found [here](gl.go).

License
-------
Apache Public License 2.0.
