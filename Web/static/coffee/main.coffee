window.log = (args...) ->
  args.splice(0, 0, new Date().format('HH:mm:ss'))
  console.log.apply(console, args)

window.Q = (value) ->
  dfd = new jQuery.Deferred()
  if value then dfd.resolve(value)
  return dfd

parseQuery = (query) ->
  result = {}
  toparse = query ? ''
  if toparse[0] is '?'
    toparse = toparse[1...]
  parts = toparse.split('&')
  for part in parts
    nv = part.split('=')
    if nv.length == 2
      result[nv[0]] = nv[1]
    else
      result[part] = yes
  return result

class DetailsDialog

  INTERVALS: [
    days: 1
    title: '1D'
  ,
    days: 3
    title: '3D'
  ,
    days: 10
    title: '10D'
  ,
    days: 28
    title: '4W'
  ,
    days: 84
    title: '12W'
  ]

  constructor: (@config) ->
    @now = new Date().getTime()
    @from = @now

  showUI: () ->
    div = $('#main-details')
    sdiv = div.find('.detail-surface').empty()
    idiv = div.find('.detail-interval').empty()
    ndiv = div.find('.detail-navigation').empty()
    makeIntervalBtn = (btn) =>
      b = @config.app.makeButton(
        text: btn.title
        target: idiv
        handler: =>
          log 'Change interval:', btn.days
      )
    for item in @INTERVALS
      makeIntervalBtn(item)
    @config.app.makeButton(
      text: 'Close'
      target: ndiv
      handler: =>
        div.hide()
    )

    div.show()

class Storage

  constructor: ->
    @query = parseQuery(window.location.search)

  get: (name, type, def) ->
    value = @query[name] ? localStorage[name] ? def
    switch (type ? 'str')
      when 'str' then return value
      when 'int'
        val = parseInt(value)
        if val is NaN then return def
        return val
      when 'bool' then return (value is 'true' or value is '1')
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
    # log 'Api Call', path, config, window.location
    p = Q()
    login = =>
      input = window.prompt('Enter Key:')
      if not input
        p.reject('Cancelled by user')
        return
      xhr(input).then( =>
        @key = input
        @storage.set(@STORAGE_KEY, input, 'str')
      )
    body = config.body ? {}
    if (config.input ? 'json') is 'json'
      dataIn = JSON.stringify(body)
    xhr = (key = @key) =>
      reqP = Q()
      $.ajax("#{@base}api/#{path}",
        contentType: 'application/json; charser=utf-8'
        data: dataIn
        dataType: config.dataOut ? 'json'
        type: 'POST'
        headers:
          'X-Key': key
        success: (data) ->
          reqP.resolve(data)
          p.resolve(data)
        error: (err) ->
          log 'Api Error:', err
          if err.status is 401
            login()
            return
          errorStr = "HTTP error: #{err.status}"
          if err.status is 403
            errorStr = err.responseText
          p.reject(err.responseText)
          reqP.reject(err.responseText)
      )
      return reqP
    xhr()
    return p

class SensorDisplay

  constructor: (@app, @config, @room) ->
    @extra = @app.parseExtra(@config.extra)

  initialize: ->
    return undefined # Render button etc here

  refresh: -> # Override
  redraw: -> # Override


window.SensorDisplay = SensorDisplay

sensorTypes = {}

window.registerSensor = (type, cls) ->
  sensorTypes[type] = cls

class AppController

  KEY_UI_DARK: 'ui_dark'
  KEY_NET_FORCE: 'net_force'
  POLL_INTERVAL_SEC: 20

  constructor: ->
    @storage = new Storage()
    @api = new APIController(@storage)
    @dark = @storage.get(@KEY_UI_DARK, 'bool', no)
    @sensors = []
    @listeners = []
    @series = []

  loadData: ->
    p = Q()
    @api.call('config').then((config) =>
      # log 'Config loaded:', config
      @makeUI(config)
    , @onError)
    return p

  toggleDark: ->
    mainTarget = $('#main')
    if @dark
      mainTarget.addClass('main-dark')
    else
      mainTarget.removeClass('main-dark')

  addDataListener: (config, handler) ->
    @listeners.push(
      config: config
      handler: handler
    )

  addSerieListener: (config, handler) ->
    @series.push(
      config: config
      handler: handler
    )

  emitDataEvent: (data) ->
    for item in @listeners
      c = item.config
      if c.device is data.device and c.index is data.index and c.type is data.type and c.measure is data.measure
        item.handler(data)

  makeNetworkControls: (menuTarget) ->
    networkBtn = @makeButton(
      icon: 'circle-o-notch'
      target: menuTarget
      handler: =>
        networkChangeHandler()
    )
    $(document).ajaxStart(=>
      networkBtn.$('i').addClass('fa-spin')
    ).ajaxStop(=>
      networkBtn.$('i').removeClass('fa-spin')
    )
    if document.webkitHidden
      eventName = 'webkitvisibilitychange'
      propName = 'webkitHidden'
    else
      eventName = 'visibilitychange'
      propName = 'hidden'
    refreshID = null
    forceRefresh = @storage.get(@KEY_NET_FORCE, 'bool', no)
    networkChangeHandler = =>
      # log 'networkChangeHandler', navigator.onLine, document[propName]
      if refreshID
        clearTimeout(refreshID)
        refreshID = null
      if (navigator.onLine and document[propName] is no) or forceRefresh
        networkBtn.almostHide(yes)
        @pollData().always(=>
          refreshID = setTimeout(=>
            networkChangeHandler()
          , @POLL_INTERVAL_SEC * 1000)
        )
      else
        networkBtn.almostHide(no)
    $(window).on('online', =>
      networkChangeHandler()
    ).on('offline', =>
      networkChangeHandler()
    )
    $(document).on(eventName, =>
      networkChangeHandler()
    )
    networkChangeHandler()

  makeUI: (config) ->
    size = $(window)
    @showError "Size: #{size.width()}x#{size.height()}"
    roomTarget = $('#main-surface')
    menuTarget = $('#main-menu')
    for item in config.layout ? []
      roomTarget.append(@makeRoom(item))
    @makeButton(
      icon: 'adjust'
      target: menuTarget
      handler: =>
        @dark = not @dark
        @storage.set(@KEY_UI_DARK, @dark, 'bool')
        @toggleDark()
    )
    @makeNetworkControls(menuTarget)
    @toggleDark()

  parseExtra: (extra) ->
    result = {}
    if not extra or not extra.length then return result
    parts = extra.split(',')
    for part in parts
      name_value = part.split('=')
      if name_value.length is 2
        val = name_value[1]
        if val[0] is '\'' and val[-1...] is '\''
          result[name_value[0]] = val[1...-1]
        else
          result[name_value[0]] = parseInt(val)
      else
        result[part] = yes
    return result

  makeButton: (config) ->
    btn = $('<button></button>').addClass('round-btn')
    if config.icon
      # Normal icon
      $("<i class='fa fa-#{config.icon}'></i>").appendTo(btn)
      btn.addClass('really-round')
    if config.cls
      btn.addClass(config.cls)
    if config.text
      $("<span class='text'></text>").appendTo(btn).text(config.text)
    if config.contents
      btn.append(config.contents)
    btn.on('click', (e) =>
      config.handler() if config.handler
    )
    if config.target
      config.target.append(btn)
    return {
      text: (text) =>
        btn.find('.text').text(text)
      html: (html) =>
        btn.html(html)
      almostHide: (visible) =>
        if visible
          btn.removeClass('almost-hidden')
        else
          btn.addClass('almost-hidden')
      '$': (arg) ->
        if arg then return btn.find(arg)
        return btn
    }

  makeRoom: (layout) ->
    # log 'Render room', layout
    wrap = $('<div></div>').addClass('room-wrap')
    div = $("""
    <div class="room">
      <div class="room-canvas"></div>
      <div class="room-side">
        <div class="room-top"></div>
      </div>
      <div class="room-side">
        <div class="room-bottom"></div>
      </div>
    </div>""")
    itemsTop = div.find('.room-top')
    itemsBottom = div.find('.room-bottom')
    wrap.append(div)
    wrap.css(
      left: "#{layout.position[0]}%"
      top: "#{layout.position[1]}%"
      width: "#{layout.position[2]}%"
      height: "#{layout.position[3]}%"
    )
    details = {}
    roomControl =
      plot: (data, colors, yaxes) =>
        log 'plot', data
        $.plot(div.find('.room-canvas'), data,
          xaxes: [
            mode: 'time'
          ]
          yaxes: yaxes ? {}
          grid:
            show: no
          colors: colors
        )
      addDetail: (name, detail) =>
        details[name] = detail
      showDetail: (name) =>
        detail = details[name]
        detail.showUI() if detail

    for sensor in layout.sensors ? []
      cls = sensorTypes[sensor.plugin]
      if not cls
        log 'Sensor type not supported', sensor
        continue
      obj = new cls(@, sensor, roomControl)
      html = obj.initialize(div)
      if html
        if sensor.revert
          itemsBottom.append(html)
        else
          itemsTop.append(html)
      @sensors.push(obj)
    return wrap

  onError: (message) =>
    @showError(message)

  showError: (message) ->
    div = $("""
    <div class="one-message"></div>
    """)
    div.text(message)
    $('#main-messages').append(div)
    setTimeout( =>
      div.remove()
    , 7000)

  fetchData: (sensors, from, to) ->
    obj =
      series: []
    for item in sensors
      obj.series.push(
        device: item.device
        type: item.type
        index: item.index
        measure: item.measure
        from: from
        to: to
      )
    return @api.call('data',
      body: obj
    ).then((data) =>
      return data
    , @onError)

  pollData: () ->
    for sensor in @sensors
      sensor.refresh()
    obj =
      sensors: []
    for item in @listeners
      obj.sensors.push(
        device: item.config.device
        type: item.config.type
        index: item.config.index
        measure: item.config.measure
      )
    return @api.call('latest',
      body: obj
    ).then((data) =>
      # log 'Data:', data
      for sensor in data.sensors
        @emitDataEvent(sensor)
    , @onError)

$(document).ready ->
  log 'App started'
  app = new AppController()
  app.loadData()

