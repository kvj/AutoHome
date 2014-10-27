window.log = (args...) ->
  args.splice(0, 0, new Date().format('HH:mm:ss'))
  console.log.apply(console, args)

class Storage

  constructor: ->

  get: (name, type, def) ->
    value = localStorage[name] ? def
    switch (type ? 'str')
      when 'str' then return value
      when 'int'
        val = parseInt(value)
        if val is NaN then return def
        return val
      when 'bool' then return (val is 'true' or val is '1')
    return value

  set: (name, value, type) ->
    val = value
    switch (type ? 'str')
      when 'int' then val = "#{value}"
      when 'bool' then val = if value then 'true' else 'false'
    localStorage[name] = val

class AppController

  constructor: ->
    @storage = new Storage()

$(document).ready ->
  log 'App started'

