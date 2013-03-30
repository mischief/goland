-- things to NOT do
--
-- dont assign newindex to userdata
--gameserver.onpacket = 'lol'
--
-- use . not : for calling go userdata
--gameserver:ANYTHING()

debug.debug()

NewItem = function(name, x, y)
  i = object.New(name)
  i.SetPos(x, y)
  i.SetTag('visible', true)
  i.SetTag('gettable', true)
  i.SetTag('item', true)
  return i
end

LoadItems = function(items)
  for k,v in pairs(items) do
    i = NewItem(v[1], v[2], v[3])

    --gameserver.LuaLog(string.format("%s %d %d", i.GetName(), i.GetPos()))

    gameserver.AddObject(i)
  end
end

items = {
  {'flag', 124, 124},
  {'flag', 124, 128},
  {'flag', 128, 128},
  {'flag', 128, 124},
}

gameserver.LuaLog("paging doctor %s", "bob")

gameserver.LoadMap('../server/map')
LoadItems(items)

