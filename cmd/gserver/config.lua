config = {
  scriptpath = os.getenv("scriptpath") or "../../scripts/?.lua",

  -- listen dialstring
  listener = os.getenv("listener") or "127.0.0.1:61507",

  factotum = {
	-- can be "tcp" too
	driver = os.getenv("factotum_driver") or "tcp",
	spec = os.getenv("factotum_spec") or "127.0.0.1:61508",
  },

  db = {
	driver = os.getenv("db_driver") or "sqlite",
	spec = os.getenv("db_spec") or "goland.db",
  },

  -- logging & debugging
  --logfile = "server.log",

  debug = os.getenv("debug") ~= nil,
  --cpuprofile = server.profile",
}

return config
