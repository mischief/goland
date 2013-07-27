-- item handling junk


-- make a new go game object and initialize it
local new = function(name, x, y)
  a = gs.Scene.Add(name .. util.IDGen())
  gs.em.Emit("newactor", a)
  pos = gs.msys.Pos()
  pos.Set(util.Pt(x, y))
  a.Add(pos)
  gs.em.Emit("propposadd", a.ID, pos)

  sp = util.NewStaticSprite(a.ID, util.NewGlyph('A', 'blue', ''))
  a.Add(sp)
  gs.em.Emit("propspriteadd", a.ID, sp)

  return a
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

