# veb: Verified (simple) Backups

Veb is styled after git, but is used for large files and/or files that do not really need version controlling, only backing up. Like:

- music
- virtual machines
- movies
- pictures

Veb isn't worth much if you only have one hard drive. It is a way to *help* you **back up** your files to another hard drive, a NAS, a thumb drive, etc.

Veb was built out of an unhealthy fear of evil hard drives and silent file corruption. You have 10GB of pictures, 50GB of music, 1TB of movies. How do you know a sector somewhere on you 10 year old hard drive hasn't gone wonky and corrupted something? How do you know you didn't accidentally overwrite (or delete) those pictures from 3 years ago? Veb gives you some tools to help, so that those automatic backups aren't automatically backing up corrupted data.

Veb tracks files via mod time, size, etc. in order to quickly determine what has changed. 'veb status' will show these changes. 

Veb also keeps a checksum of every file as it is added to the veb repository. 'veb verify' will re-check every file to see if its contents have been silently changed. Verification could take some time... It has to read everything in the repository, so if the repository is 1TB, you may want to get a snack while veb does the math.

What veb won't do is read your mind. You'll have to remember what files you've changed so when you run 'veb status' or 'veb verify', you can parse the results and fix or commit as needed.


## veb commands:

    init   - initializes a new veb repository at the current directory
    status - quick check of what's new or changed, no recomputing of checksums
    verify - slow check of all files, recomputing all checksums
    commit - blesses all new/changed files as good & adds them to the repository
    remote - sets the backup location for this repository
    push   - sends committed files in current (local) repository to remote repo
             only sends files the remote doesn't have the latest of
    pull   - gets committed files from remote repo
             only gets files the local repo doesn't have the latest of
    sync   - veb pull & veb push
    fix    - pulls the specified file from the remote, overwriting the local copy
    help   - prints help


## TODOs and planned features

This is veb v0.1, so a lot is still to come.

- Only init, status, verify, commit, remote, and push currently work
  - These represent the minimal working set of commands, so it's a good spot to drop a v0.1 tag.
  - pull, sync, fix and help will follow shortly
- Deleted files currently just hang around in the index. They will be report in 'veb status'/'veb verify', and removed from the repository's index as part of 'veb commit'.
- Nice: veb currently runs at default priority. You can nice it yourself (e.g. 'nice veb push'), but for something that's doing so much file IO, it should be niced by default.
- Actual remote repos: veb currently can only work on mounted filesystems. Over-the-network remotes are planned.
  - Also planned: rsync or equivalent for push/pull instead of current "copy the whole thing all over again".
- Reduce package main's footprint: A lot of work currently happens in veb/veb.go. This will all be moved into the veb/veb package so that veb.go is the lightweight user interface, and all work happens in the actual veb library.
- Choice of hash function: Currently SHA1 is hard-coded. Plan is to allow at least SHA1, SHA256, and MD5 during 'veb init'.
  - MD5 may be useful for people who have huge files (Virtual Machines, for example) and need fast hashing.
- Testing: Will be added. Have been white-box testing to this point, but need actual test suites going forward.
- Also, a sprinkling of TODOs in the code need to be TODONE.

## veb's intended usage

1. Find some things you want veb to care about. 
2. Initialize a veb repository there ('veb init') & commit those files ('veb commit'). 
3. Do the same for the remote (you only need 'veb init' if your remote is a fresh, new, empty folder).
4. Tell your (first) veb repo where the remote you just created is ('veb remote /path/to/remote').
5. Back up ('veb push').
6. If you're interested in what's changed since you last backed up, 'veb status' will tell you.
7. You should 'veb verify' every once in a while to see if anything got corrupted. (Week? Month? Whatever you're comfortable with.)
8. 'veb status', 'veb commit' and 'veb push' regularly to get your latest data backed up.

Just be sure to pay attention to 'veb status' and 'veb verify' output. Veb can only tell you what's changed; it doesn't know if a change means you modified that mp3 or if that mp3 got corrupted.

## veb metadata

Veb keeps all its information in a folder called .veb, located in whatever folder you ran 'veb init' from.

It contains three files:

- .veb/index
- .veb/xsums
- .veb/log.txt

The index is an index off your committed files, encoded in Go's [gob](http://blog.golang.org/2011/03/gobs-of-data.html) format.

xsums is a md5sum/sha1sum/shasum formatted file. If you want to test veb's checksumming sanity, you can use that as the checkfile for those tools.

    palladium:local spydez$ shasum -c .veb/xsums
      (many lines of shasum saying OK go here)

log.txt is a plain text file containing info & error logs from all your veb commands.

    palladium:local spydez$ cat .veb/log.txt 
    info  >> 2012/05/21 23:30:19 log.go:42: ENTERING commit
    info  >> 2012/05/21 23:30:21 veb.go:619: commit (39 commits, 0 errors) took 1.553082s
    info  >> 2012/05/21 23:30:21 log.go:47: LEAVING  commit
    info  >> 2012/05/21 23:30:21 veb.go:202: done
    
    
    info  >> 2012/05/21 23:31:05 log.go:42: ENTERING remote
    error >> 2012/05/21 23:31:05 veb.go:655: stat /Users/spydez/sourcepan/go/src/spydez/veb/test/scratch/remote/.veb: no such file or directory
    info  >> 2012/05/21 23:31:05 log.go:47: LEAVING  remote
    info  >> 2012/05/21 23:31:30 log.go:42: ENTERING remote
    info  >> 2012/05/21 23:31:30 veb.go:679: remote took 1.499ms
    info  >> 2012/05/21 23:31:30 log.go:47: LEAVING  remote
    info  >> 2012/05/21 23:31:30 veb.go:202: done

The .veb folder, and everything in it, are ignored by veb commands. It does copy index and xsums to index~ and xsums~ before writing new ones, for some rudimentary self-backing up.

## A short veb unguided tour
    palladium:scratch spydez$ cd local

    palladium:local spydez$ veb init
    Initialized empty veb repository at ~/test/scratch/local

    palladium:local spydez$ veb commit
    veb repository at ~/test/scratch/local 
    
    ----------------
    Committed files:
    ----------------
      .DS_Store
      311/Don't Tread On Me/09 Whiskey & Wine.test.bin
      311/Don't Tread On Me/10 It's Getting OK Now.test.bin
      311/Don't Tread On Me/02 Thank Your Lucky Stars.test.bin
      311/Don't Tread On Me/11 There's Always An Excuse.test.bin
      Fall Out Boy/Folie À Deux/03 She's My Winona.test.bin
      Ace Troubleshooter/It's Never Enough/02 Anything.test.bin
      Fall Out Boy/Folie À Deux/01 Disloyal Order Of Water Buffaloes.test.bin
      311/Don't Tread On Me/01 Don't Tread On Me.test.bin
      Apt° Core/2/01 No Such Thing As Time.test.bin
      Ace Troubleshooter/It's Never Enough/01 Ball & Chain.test.bin
      Fall Out Boy/Folie À Deux/02 I Don't Care.test.bin
      Fall Out Boy/Folie À Deux/04 America's Suitehearts.test.bin
      Five Iron Frenzy/Our Newest Album Ever!/01 Handbook for the Sellout.test.bin
      Fall Out Boy/Folie À Deux/05 Headfirst Slide Into Cooperstown On A Bad Bet.test.bin
      Five Iron Frenzy/Our Newest Album Ever!/11 Oh, Canada.test.bin
      Movits!/Äppelknyckarjazz/01 Ta på dig dansskorna.test.bin
      Five Iron Frenzy/Our Newest Album Ever!/02 Where is Micah_.test.bin
      Five Iron Frenzy/Our Newest Album Ever!/13 Every New Day.test.bin
      Five Iron Frenzy/Our Newest Album Ever!/12 Most Likely to Succeed.test.bin
      Movits!/Äppelknyckarjazz/06 Fel del av gården.test.bin
      Movits!/Äppelknyckarjazz/03 Swing för hyresgästföreningen.test.bin
      Movits!/Äppelknyckarjazz/04 Fast tvärtom.test.bin
      Movits!/Äppelknyckarjazz/08 Tom Jones.test.bin
      Movits!/Äppelknyckarjazz/09 Äppelknyckarjazz.test.bin
      Movits!/Äppelknyckarjazz/11 2 dollar på fickan.test.bin
      bar.txt
      foo.txt
      Movits!/Äppelknyckarjazz/10 Stick iväg Jack del II.test.bin
      Parov Stelar/Coco Pt.1/01 Coco (Featuring Lilja Bloom).test.bin
      Parov Stelar/Coco Pt.2/03 Silent Snow(Featuring Max The Sax).test.bin
      Parov Stelar/Coco Pt.2/02 Ragtime Cat(Featuring Lilja Bloom).test.bin
      Parov Stelar/Coco Pt.2/01 The Mojo Radio Gang(Radio Ver.).test.bin
      Parov Stelar/Coco Pt.1/02 Hurt.test.bin
      Trans-Siberian Orchestra/Christmas Eve and Other Stories/05 The Silent Nutcracker (Instrumental).test.bin
      Parov Stelar/Coco Pt.1/03 For Rose(수원 아이파크 시티 CF삽입곡).test.bin
      Parov Stelar/Coco Pt.2/04 Libella Swing(현대카드 TV CF 삽입곡).test.bin
      Trans-Siberian Orchestra/Christmas Eve and Other Stories/06 A Mad Russian's Christmas (Instrumental).test.bin
      Trans-Siberian Orchestra/Christmas Eve and Other Stories/02 O Come All Ye Faithful_O Holy Night (instrumental).test.bin
    
    summary: 39 commits, 0 errors in 1.553082s

    palladium:local spydez$ mkdir ../remote
    palladium:local spydez$ cd ../remote

    palladium:remote spydez$ veb init
    Initialized empty veb repository at ~/test/scratch/remote

    palladium:remote spydez$ cd ../local

    palladium:local spydez$ veb remote ../remote
    veb repository at ~/test/scratch/local 
    
    veb added ~/test/scratch/remote as the remote

    palladium:local spydez$ veb push
    veb repository at ~/test/scratch/local 
    
                                                                                    
    -------------
    Pushed files:
    -------------
      311/Don't Tread On Me/10 It's Getting OK Now.test.bin                         
      Movits!/Äppelknyckarjazz/09 Äppelknyckarjazz.test.bin                         
      Fall Out Boy/Folie À Deux/02 I Don't Care.test.bin                            
      foo.txt                                                                       
      Parov Stelar/Coco Pt.1/03 For Rose(수원 아이파크 시티 CF삽입곡).test.bin      
      Ace Troubleshooter/It's Never Enough/02 Anything.test.bin                     
      Fall Out Boy/Folie À Deux/04 America's Suitehearts.test.bin                   
      Parov Stelar/Coco Pt.2/03 Silent Snow(Featuring Max The Sax).test.bin         
      Fall Out Boy/Folie À Deux/05 Headfirst Slide Into Cooperstown On A Bad Bet.test.bin
      bar.txt                                                                       
      Movits!/Äppelknyckarjazz/06 Fel del av gården.test.bin                        
      Trans-Siberian Orchestra/Christmas Eve and Other Stories/02 O Come All Ye Faithful_O Holy Night (instrumental).test.bin
      Fall Out Boy/Folie À Deux/01 Disloyal Order Of Water Buffaloes.test.bin       
      Fall Out Boy/Folie À Deux/03 She's My Winona.test.bin                         
      Parov Stelar/Coco Pt.1/02 Hurt.test.bin                                       
      Five Iron Frenzy/Our Newest Album Ever!/02 Where is Micah_.test.bin           
      .DS_Store                                                                     
      Ace Troubleshooter/It's Never Enough/01 Ball & Chain.test.bin                 
      Five Iron Frenzy/Our Newest Album Ever!/13 Every New Day.test.bin             
      Parov Stelar/Coco Pt.2/04 Libella Swing(현대카드 TV CF 삽입곡).test.bin       
      Movits!/Äppelknyckarjazz/03 Swing för hyresgästföreningen.test.bin            
      Five Iron Frenzy/Our Newest Album Ever!/12 Most Likely to Succeed.test.bin    
      Movits!/Äppelknyckarjazz/10 Stick iväg Jack del II.test.bin                   
      Movits!/Äppelknyckarjazz/04 Fast tvärtom.test.bin                             
      Five Iron Frenzy/Our Newest Album Ever!/11 Oh, Canada.test.bin                
      311/Don't Tread On Me/02 Thank Your Lucky Stars.test.bin                      
      Movits!/Äppelknyckarjazz/01 Ta på dig dansskorna.test.bin                     
      Trans-Siberian Orchestra/Christmas Eve and Other Stories/05 The Silent Nutcracker (Instrumental).test.bin
      Apt° Core/2/01 No Such Thing As Time.test.bin                                 
      Trans-Siberian Orchestra/Christmas Eve and Other Stories/06 A Mad Russian's Christmas (Instrumental).test.bin
      311/Don't Tread On Me/11 There's Always An Excuse.test.bin                    
      Five Iron Frenzy/Our Newest Album Ever!/01 Handbook for the Sellout.test.bin  
      311/Don't Tread On Me/01 Don't Tread On Me.test.bin                           
      Movits!/Äppelknyckarjazz/08 Tom Jones.test.bin                                
      Parov Stelar/Coco Pt.1/01 Coco (Featuring Lilja Bloom).test.bin               
      Parov Stelar/Coco Pt.2/01 The Mojo Radio Gang(Radio Ver.).test.bin            
      311/Don't Tread On Me/09 Whiskey & Wine.test.bin                              
      Parov Stelar/Coco Pt.2/02 Ragtime Cat(Featuring Lilja Bloom).test.bin         
      Movits!/Äppelknyckarjazz/11 2 dollar på fickan.test.bin                       
                                                                                    
    status:    0 ignored,    0 errors,   39 pushed,    0 unchanged in 5.383501s

    palladium:local spydez$ touch bar.txt 
    palladium:local spydez$ echo foo > foo.txt

    palladium:local spydez$ veb status
    veb repository at ~/test/scratch/local 
    
    --------------
    Changed files:
    --------------
      bar.txt
          - modified on (2012-05-21 23:31:47 -0500 CDT)
    
      foo.txt
          - filesize increased 4.00B (0.00B -> 4.00B)
          - modified on (2012-05-21 23:32:02 -0500 CDT)
    
    
    MAKE SURE CHANGED FILES ARE THINGS YOU'VE ACTUALLY CHANGED
      (use 'veb fix <file>' if a file has been corrupted in this repository)
      (use 'veb push', 'veb pull', or 'veb sync' to commit changed/new files)
    
    summary: 0 new, 2 changed (2.647ms)

    palladium:local spydez$ veb push
    veb repository at ~/test/scratch/local 
    
    use 'veb status' to check new/changed files
    use 'veb commit' to add new/changed files to repository
    
    --------------------
    LOCAL ignored files:
    --------------------
      bar.txt
      foo.txt
                                                                                    
    status:    2 ignored,    0 errors,    0 pushed,   37 unchanged in 10.196ms

    palladium:local spydez$ veb commit
    veb repository at ~/test/scratch/local 
    
    ----------------
    Committed files:
    ----------------
      bar.txt
      foo.txt
    
    summary: 2 commits, 0 errors in 3.222ms

    palladium:local spydez$ veb push
    veb repository at ~/test/scratch/local 
    
                                                                                    
    -------------
    Pushed files:
    -------------
      foo.txt                                                                       
                                                                                    
    status:    0 ignored,    0 errors,    1 pushed,   38 unchanged in 4.626ms
