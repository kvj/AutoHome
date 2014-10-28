window.log = (args...) ->
  args.splice(0, 0, new Date().format('HH:mm:ss'))
  console.log.apply(console, args)

window.Q = (value) ->
  dfd = new jQuery.Deferred()
  if value then dfd.resolve(value)
  return dfd

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

class APIController

  STORAGE_KEY: 'api_key'

  constructor: (@storage, @base = '') ->
    @key = @storage.get(@STORAGE_KEY, 'str', '')

  call: (path, config = {}) ->
    log 'Api Call', path, config
    p = Q()
    login = =>
      input = window.prompt('Enter Key:')
      xhr(input).then( =>
        @key = input
        @storage.set(@STORAGE_KEY, input, 'str')
      )
    body = config.body ? {}
    if (config.input ? 'json') is 'json'
      dataIn = JSON.stringify(body)
    xhr = (key = @key) =>
      $.ajax("#{@base}api/#{path}",
        contentType: 'application/json; charser=utf-8'
        data: dataIn
        dataType: config.dataOut ? 'json'
        type: 'POST'
        headers:
          'X-Key': key
        success: (data) ->
          p.resolve(data)
        error: (err) ->
          log 'Api Error:', err
          if err.status is 401
            login()
            return
          p.reject("HTTP error: #{err.status}")
      )
      return p
    xhr()
    return p

class AppController

  constructor: ->
    @storage = new Storage()
    @api = new APIController(@storage)

  loadData: ->
    p = Q()
    @api.call('config').then((config) =>
      log 'Config loaded:', config
    , @onError)
    return p

  onError: (message) ->
    alert(message)

$(document).ready ->
  log 'App started'
  app = new AppController()
  app.loadData()

