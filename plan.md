#Home page

- have search bar can search for song, artist, playlist
- by default show random 5 playlists (public one), each playlist show only 5song, (have detail for playlist which show all songs in that playlist)
- the song in the list of playlist no need detail, only need name, artist, id

#Playlist detail page

- list all songs in the playlist
- the song in the list of playlist no need detail, only need name, artist, id
- in playlist as song have covers, the playlist able to select certain cover for the song as default to play in the playlist
  - so there should be sth to mark e.g song A default cover B for playlist X

#Song detail page

- artistsType
  - name string
  - cv artistsType

- coverType
  - id string (will use youtube video id for this)
  - artists artistsType[]

- have all detail about the song
  - id string (will use youtube video id for this)
  - title string
  - altTitle string[]
  - artists artistsType[]
  - covers coverType[]
  - lyrics (have support for multiple language e.g. english, japanese, romaji, chinese, pinyin, etc)

#Add song page

- have form that input what in song information, cover is optional

#user page

- show own playlist include both public and private playlist
