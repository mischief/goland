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
    i.SetGlyph(util.NewGlyph(v[4]))

    --gameserver.LuaLog(string.format("%s %d %d", i.GetName(), i.GetPos()))

    gs.AddObject(i)
  end
end

items = {
  {'flag', 124, 124, '⚑'},
  {'flag', 124, 128, '⚑'},
  {'flag', 128, 128, '⚑'},
  {'flag', 128, 124, '⚑'},
  {'block', 125, 123, '☒'},
  {'block', 126, 125, '☒'},
  {'block', 127, 128, '☒'},
  {'block', 128, 127, '☒'},
}

gs.LoadMap('../server/map')
LoadItems(items)

collide = function(o1, o2)
  gs.LuaLog("Colliding! %s %s", o1.GetName(), o2.GetName())

  if o2.GetName() == 'block' then
    x1,y1 = o1.GetPos()
    gs.LuaLog("o1 %d %d", x1, y1)
    x2,y2 = o2.GetPos()
    gs.LuaLog("o2 %d %d", x2, y2)

    if x1 ~= x2 or y1 ~= y2 then
      newx = x2-(x1-x2)
      newy = y2-(y1-y2)
      gs.LuaLog("%s move to %f %f", o2.GetName(), newx, newy)
      o2.SetPos(newx, newy)
    end
  end
end

