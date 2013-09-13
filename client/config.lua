-- goland client configuration

config = {
  -- username
  username    = os.getenv("USER"), -- "anonymous" .. math.random(1, 256),

  -- server
  server      = "127.0.0.1:61507",
  --server      = "iota.offblast.org:61507",

  -- logging & debugging
  --logfile     = "client.log",

  debug       = "true",
  --cpuprofile  = "client.profile",

  -- keybindings
  keys = {
    -- player control
    w = "moveup",
    a = "moveleft",
    s = "movedown",
    d = "moveright",

    k = "moveup",
    h = "moveleft",
    j = "movedown",
    l = "moveright",

    [","] = "pickup",
    x = "drop",
    i = "inventory",

    -- gui control
    esc = "quit",
    enter = "chat",
    ["`"]   = "console",
  },

  -- enable/disable console commands
  ---[[
  commands = {
    exec = false,
    lua = false,
  },
  --]]

  -- set theme, optional
  ---[[
  theme = {
    titlefg = "magenta",
    borderfg = "blue",
    textfg = "white",

    promptfg = "cyan",
  }
  --]]

}

return config

