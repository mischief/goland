-- item handling junk


-- make a new go game object and initialize it
local new = function(name, x, y)
  i = object.New(name)
  i.SetPos(x, y)
  i.SetTag('visible', true)
  i.SetTag('gettable', true)
  i.SetTag('item', true)
  return i
end

-- load a table of items like
-- { {'flag', 2, 4, '4'}, ... }
local load = function(items)
  for k,v in pairs(items) do
    i = new(v[1], v[2], v[3])
    i.SetGlyph(util.NewGlyph(v[4], v[5], v[6]))

    --gameserver.LuaLog(string.format("%s %d %d", i.GetName(), i.GetPos()))

    gs.AddObject(i)
  end
end

return {
  new = new,
  load = load,
}

