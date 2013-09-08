-- loot database

function file_load(location, filename)
    local path = location .. "/" .. filename
    local f = assert(io.open(path, "r"))
    local c = f:read "*a"

    f:close()

    return c
end

local file_contents = file_load('../game/data', 'loot.json')
local DB = Json.Decode(file_contents)--file_load('../game/data', 'loot.json')) -- decode example

local make = function(itemid, posx, posy)
    i = DB[itemid]
    return { i.name, posx, posy, i.glyph, i.color_fg, i.color_bg }
end

local colorize = function(item, color)
    item[5] = color
    return item
end

return {
    make = make,
    colorize = colorize
}
