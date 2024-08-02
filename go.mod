module m6kparse

go 1.18

require (
	github.com/google/gopacket v1.1.19
	github.com/quarkslab/wirego/wirego v0.0.0-20240403125855-b964e68cd018
	gitlab.qb/bgirard/wirego/wirego v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.0.0-20190412213103-97732733099d // indirect

replace gitlab.qb/bgirard/wirego/wirego => /Users/benoit/Documents/Dev/wirego/wirego
