#Home page

- have search bar can search for song, artist, playlist
- by default show random 5 playlists (public one), each playlist show only 5song, (have detail for playlist which show all songs in that playlist)
- the song in the list of playlist no need detail, only need name, artist, id

#Playlist detail page

- list all songs in the playlist
- the song in the list of playlist no need detail, only need name, artist, id

#Song detail page

- artistsType
  - name string

- have all detail about the song
  - id string (will use youtube video id for this)
  - name string []
  - artists artistsType[]
  - covers
    - artists artistsType[]
    - id string (will use youtube video id for this)
  - lyrics (have support for multiple language e.g. english, japanese, romaji, chinese, pinyin, etc)

#Add song page

- have form that input what in song information, cover is optional

#user page

- show own playlist include both public and private playlist
