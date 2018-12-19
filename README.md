# md-sleep
md-sleep: watch md-raid array and spin down idle disks

md-sleep will check `/sys/block/mdX/stats` constantly for disk-i/o and
will spin up/down the slave disks in your raid array according to your
preferences. The program uses `hdparm` to interact with the disks.