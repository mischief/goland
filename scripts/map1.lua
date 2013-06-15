-- map1.lua - first map yay

collision = require('collision')

-- the goods
flags = {
  {'flag', 124, 124, '⚑'},
  {'flag', 124, 128, '⚑'},
  {'flag', 128, 128, '⚑'},
  {'flag', 128, 124, '⚑'},
}

blocks = {
  {'block', 125, 123, '☒'},
  {'block', 126, 125, '☒'},
  {'block', 127, 128, '☒'},
  {'block', 128, 127, '☒'},
}

score = {
  {'scorepoint', 97, 115, '_'},
  {'scorepoint', 124, 130, '_'},
}

nflags = #flags

scores = {}

collision.fns.scorepoint = function(o1, o2)
  if o2.GetTag('player') ~= true then
    return false
  end

  gs.LuaLog("%s has stepped on a %s", o2.GetName(), o1.GetName())

  subobjs = o2.SubObjects
--  print(type(subobjs))
  slice = subobjs.GetSlice()
--  print(type(slice), #slice)
--  for k,v in pairs(getmetatable(slice)) do
--    print(k,v)
--  end
--  obj = slice[1]
--  print(type(obj))
  slice = assert(slice:Slice(0, #slice)) -- XXX why does this work????
--  print('slice has ', #slice)

  for i=2, #slice do
    obj = slice[i]
--    print(type(obj))

    -- if the player has a flag, take it and give them a point
    -- then spawn another flag
    if obj.GetName() == 'flag' then

      gs.LuaLog("%s has flag, removing", o2.GetName())
      o2.RemoveSubObject(obj)


      pname = o2.GetName()
      if scores[pname] ~= nil then
        scores[pname] = scores[pname] + 1
      else
        scores[pname] = 1
      end

      gs.SendPkStrAll("Rchat", string.format("ctf: %s has %d points!", pname, scores[pname]))

      -- now spawn a new flag
      gs.LuaLog("nflags %.0f", nflags)
      nflags = nflags - 1
      gs.LuaLog("nflags %.0f", nflags)

      if nflags == 2 then
        f = flags[math.random(1,4)]
        items.load({f})
        nflags = nflags + 1
        gs.LuaLog("nflags %.0f", nflags)
      end


    end
  end

  return true
end

fns = {}

fns.load = function()
  gs.LoadMap('../server/map')
  items.load(flags)
  items.load(blocks)

  -- load the score point
  for _, v in pairs(score) do
    sp = object.New(v[1])
    sp.SetPos(v[2], v[3])
    sp.SetTag('visible', true)
    sp.SetGlyph(util.NewGlyph(v[4]))
    gs.AddObject(sp)
  end
end

return fns

