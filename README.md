[![Build Status](https://travis-ci.org/TcM1911/jamsonic.svg?branch=master)](https://travis-ci.org/TcM1911/jamsonic) [![Go Report Card](https://goreportcard.com/badge/github.com/TcM1911/jamsonic)](https://goreportcard.com/report/github.com/TcM1911/jamsonic)
# Jamsonic

Jamsonic is a console based music player for Subsonic/Madsonic/Libresonic/Airsonic.
It's forked from [Jam](https://github.com/budkin/jam), which is a console based
music player for Google Play Music.

It has the following features:

- Populating the database with artists and albums you saved through the
  web interface (or by any other means)
- Playing, pausing, stopping, previous track, next track

Contributions are welcome!

## How to get up and running?

For macOS and Linux, portaudio has to be installed. Windows doesn't need anything extra.

## Keybindings

The keybindings are mostly the same as in Cmus:

| Key           | Action                                                                       |
|---------------|------------------------------------------------------------------------------|
| return, x     | play currently selected artist, album or song                                |
| c             | pause                                                                        |
| v             | stop                                                                         |
| b             | next track                                                                   |
| z             | previous track                                                               |
| Ctrl+u        | synchronize the database (in case you added some songs in the web interface) |
| /             | search artists                                                               |
| n             | next search result                                                           |
| tab           | toggle artists/tracks view                                                   |
| escape        | quit                                                                         |
| up arrow, k   | scroll up                                                                    |
| down arrow, j | scroll down                                                                  |
| Home, g       | scroll to top                                                                |
| End, G        | scroll to bottom                                                             |
| space         | toggle albums                                                                |
| R             | randomize artists                                                            |
| Ctrl+Space    | toggle view (playlists/artists)                                              |
| r             | repeat current track                                                         |
