Lightswitch Rave - a simple, fast web frontend to your music library for Linux
==============================================================================

Lightswitch Rave is a way to easily share a music collection over the
world wide web.  It provides a dead-simple web application enabling
authorized users to browse your music by artist or album.  Songs are
played using the 'html5' audio tag.  This means that support for
mp3/aac/ogg audio is dependent on your browser.  Firefox doesn't come
with the codecs for mp3 or aac, so unless your music library is all
ogg, use Chrome

I originally designed it so that I could access the music I had on my
desktop at home from my laptop at work, and at this point it performs
the task admirably.

![Lightswitch Rave in chrome](https://github.com/bpowers/lightswitchrave/raw/master/doc/cnote_in_use.png "Lightswitch Rave in Chrome")

Lightswitch Rave uses inotify to watch for new/changed files, so it is
currently linux only.  kqueue provides similar functionality on
Mac/BSD, so abstracting this out would be possible, but I don't have
plans to do that right now.


highlights
----------

- A sweet login page.


status
------

It is somewhat a work in progress.  I had a previous project
[cnote](https://github.com/bpowers/cnote/) which provided the
API/sqlite indexing of my music library.  In order to easily add new
features like authentication and uploading, I've switched to Go.
However, I haven't finished the music indexing part, so lightswitch
proxies /api to cnote.  With this said, the two of them running
together work excellently.


license
-------

Lightswitch Rave is offered under the MIT license, see COPYING for details.
