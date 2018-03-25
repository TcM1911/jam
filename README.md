### This project is no longer maintained as I've made the switch from GPM to Deezer and I can't test it. If you want to take over, drop me an email

# jamsonic

This is my first Go program, I wanted to listen to Google Play Music on console,
so I've written a player. It is inspired by Matt Jibson's [Moggio](https://github.com/mjibson/moggio/) and uses one of
his libraries. You can see it in action if you follow this link:
https://www.dropbox.com/s/ygai22mzgmtd2ri/out-2.ogv

It has the following features:

- Last.fm scrobbling (use -lastfm flag)
- populating the database with artists and albums you saved through the
  web interface (or by any other means)
- searching within artists in the database
- playing, pausing (buggy, I need help with it), stopping, previous track, next
  track
- the interface is ripped off from Cmus, I only added a progress bar
- this player no longer lists artists in random order - if you want to shuffle
  them press R

In order to use this program you must be logged in in Google Play services on
your phone, if you have no smartphone - this program, at its current state,
is not for you

If you use 2-factor authorization in your Google account, you must
generate an app password, follow this link 
https://security.google.com/settings/security/apppasswords

The linux binary I release is not static, it depends on pulseaudio, if you want
to build it from source, you are going to need the pulseaudio development package
installed.
Windows users are all set



If you have an x86 system, you'll have to compile it yourself, sorry for that

Contributions are welcome!

The keybindings are mostly the same as in Cmus:

| Key           | Action                                                                       |
|---------------|------------------------------------------------------------------------------|
| return, x     | play currently selected artist, album or song                                |
| c             | pause                                                                        |
| v             | stop                                                                         |
| b             | next track                                                                   |
| z             | previous track                                                               |
| u             | synchronize the database (in case you added some songs in the web interface) |
| /             | search artists                                                               |
| n             | next search result                                                           |
| tab           | toggle artists/tracks view                                                   |
| escape, q     | quit                                                                         |
| up arrow, k   | scroll up                                                                    |
| down arrow, j | scroll down                                                                  |
| Home, g       | scroll to top                                                                |
| End, G        | scroll to bottom                                                             |
| space         | toggle albums                                                                |
| R             | randomize artists                                                            |
| Ctrl+Space    | toggle view (playlists/artists)                                              |
| r             | repeat current track                                                         |

[1]: https://github.com/mjibson/moggio



TODO
- make the interface detachable (like MOC)
- make the binary able to receive command line arguments for controlling playback
  (next track, pause, etc)
- implement search within the GPM global database
- feature requests are welcome as well


Contributors (in order of the number of commits):
- @nlamirault
- @koron
- @avilanicolas
- @felixonmars
- @MartijnBraam
