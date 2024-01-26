---
--- Generated by EmmyLua(https://github.com/EmmyLua)
--- Created by Administrator.
--- DateTime: 2024/1/11 11:33
---
local key = KEYS[1]
local cntKey = key .. ":cnt"
local expectedVal = ARGV[1]
local cnt = tonumber(redis.call("get", cntKey))

if cnt == nil or cnt <= 0 then
    return -1
end
local val = redis.call("get", key)
if val ~= expectedVal then
    redis.call("decr", cntKey)
    return -2
end
return 0
