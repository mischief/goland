config = {
  -- username
  username    = os.getenv("USER"), -- "anonymous" .. math.random(1, 256),

  -- server
  server      = "127.0.0.1:61507",

  -- logging & debugging
  logfile     = "client.log",

  debug       = "true",
  --cpuprofile  = "client.profile",

}

return config

