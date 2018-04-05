[![Build Status](https://travis-ci.org/TcM1911/jamsonic.svg?branch=master)](https://travis-ci.org/TcM1911/jamsonic)
# Jamsonic

Jamsonic is a console based music player for Subsonic/Madsonic/Libresonic/Airsonic.
It's forked from [Jam](https://github.com/budkin/jam), which is a console based
music player for Google Play Music.

It has the following features:

- Google Play Music is still supported. (use -googlemusic flag)
- Last.fm scrobbling (use -lastfm flag)
- populating the database with artists and albums you saved through the
  web interface (or by any other means)
- searching within artists in the database
- playing, pausing (buggy), stopping, previous track, next track
- the interface is ripped off from Cmus, I only added a progress bar
- this player no longer lists artists in random order - if you want to shuffle
  them press R

Contributions are welcome!

## How to get up and running?

The linux binaries released depends on pulseaudio. For macOS, portaudio has to be installed. Windows doesn't need anything extra.

## Keybindings

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

## Google Play Music

In order to use this program you must be logged in in Google Play services on
your phone, if you have no smartphone - this program, at its current state,
is not for you

If you use 2-factor authorization in your Google account, you must
generate an app password, follow this link 
https://security.google.com/settings/security/apppasswords
