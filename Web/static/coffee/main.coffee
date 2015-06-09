window.log = (args...) ->
  args.splice(0, 0, new Date().format('HH:mm:ss'))
  console.log.apply(console, args)

window.Q = (value) ->
  dfd = new jQuery.Deferred()
  if value then dfd.resolve(value)
  return dfd

window.Q.all = (arr) ->
  return jQuery.when.apply(jQuery, arr)

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


class Dialog
  
  AUTO_CLOSE_TIMEOUT: 180

  constructor: (@config) ->
    @loading = 0
    @closeInterval = if @config.autoClose >= 0 then @config.autoClose else @AUTO_CLOSE_TIMEOUT

  isLoading: ->
    return @loading > 0
  isVisible: ->
    return @visible

  toggleLoading: (stop = no) ->
    if stop
      @loading--
    else
      @loading++
    if @loading is 1
      $('#main-details').find('.detail-indicator').removeClass('no-display')
    if @loading is 0
      $('#main-details').find('.detail-indicator').addClass('no-display')

  showUI: () ->
    div = $('#main-details')
    sdiv = div.find('.detail-surface').empty()
    idiv = div.find('.detail-interval').empty()
    ndiv = div.find('.detail-navigation').empty()
    @config.app.makeButton(
      icon: 'close'
      target: ndiv
      handler: =>
        @closeDialog()
    )
    @visible = yes
    div.removeClass('no-display')
    @config.onCreate(sdiv, ndiv) if @config.onCreate
    @scheduleAutoClose()
  
  closeDialog: ->
    @visible = no
    div = $('#main-details')
    div.addClass('no-display')
    clearTimeout(@closeTimeoutID) if @closeTimeoutID
    @config.onClose() if @config.onClose

  scheduleAutoClose: ->
    clearTimeout(@closeTimeoutID) if @closeTimeoutID
    if @closeInterval is 0 then return
    @closeTimeoutID = setTimeout(=>
      @closeDialog()
    , 1000 * @closeInterval)

  addSurface: (d) ->
    div = $('#main-details')
    sdiv = div.find('.detail-surface')
    sdiv.append(d)
    return {
      div: d
    }


class DetailsDialog

  AUTO_CLOSE_TIMEOUT: 180

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
    @loading = 0

  toggleLoading: (stop = no) ->
    if stop
      @loading--
    else
      @loading++
    if @loading is 1
      $('#main-details').find('.detail-indicator').removeClass('no-display')
    if @loading is 0
      $('#main-details').find('.detail-indicator').addClass('no-display')

  plot: (div, data, colors, yaxes) ->
    # log 'plot', data
    $.plot(div, data,
      xaxes: [
        mode: 'time'
        timezone: 'browser'
      ]
      yaxes: yaxes ? {}
      grid:
        show: yes
      colors: colors
    )

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
          # log 'Change interval:', btn.days
          @changeInterval(btn.days)
      )
    for item in @INTERVALS
      makeIntervalBtn(item)
    @days = @config.default ? 1
    @config.app.makeButton(
      icon: 'chevron-left'
      target: ndiv
      handler: =>
        @moveDate(-1)
    )
    @config.app.makeButton(
      icon: 'circle-o'
      target: ndiv
      handler: =>
        @moveDate(0)
    )
    @config.app.makeButton(
      icon: 'chevron-right'
      target: ndiv
      handler: =>
        @moveDate(1)
    )
    @scheduleAutoClose()
    @config.app.makeButton(
      icon: 'close'
      target: ndiv
      handler: =>
        @closeDialog()
    )
    @visible = yes
    @now = new Date().getTime()
    div.removeClass('no-display')
    @config.onCreate(sdiv) if @config.onCreate
    @moveDate(0)
  
  closeDialog: ->
    @visible = no
    div = $('#main-details')
    div.addClass('no-display')
    clearTimeout(@closeTimeoutID)

  scheduleAutoClose: ->
    clearTimeout(@closeTimeoutID) if @closeTimeoutID
    @closeTimeoutID = setTimeout(=>
      @closeDialog()
    , 1000 * @AUTO_CLOSE_TIMEOUT)

  addSurface: (d) ->
    div = $('#main-details')
    sdiv = div.find('.detail-surface')
    sdiv.append(d)
    return {
      div: d
    }

  changeInterval: (days) ->
    @days = days
    if @config.forecast
      @to = new Date(@from.getTime())
      @to.setDate(@to.getDate()+@days)
    else
      @from = new Date(@to.getTime())
      @from.setDate(@from.getDate()-@days)
    @config.onRender() if @config.onRender
    @scheduleAutoClose()

  moveDate: (dir = 0) ->
    switch dir
      when 0
        @from = new Date(@now)
        @to = new Date(@now)
        if @config.forecast
          @to.setDate(@to.getDate()+@days)
        else
          @from.setDate(@from.getDate()-@days)
      when 1
        @to.setDate(@to.getDate()+@days)
        @from.setDate(@from.getDate()+@days)
      when -1
        @to.setDate(@to.getDate()-@days)
        @from.setDate(@from.getDate()-@days)
    # log 'moveDate', @from, @to
    @config.onRender() if @config.onRender
    @scheduleAutoClose()

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

$.ajaxTransport('raw', (opt, orig_opt, jqxhr) ->
  return {
    send: (headers, callback) ->
      xhr = new XMLHttpRequest()
      dataType = opt.responseType ? 'blob'
      xhr.addEventListener('load', ->
        data = {}
        data[opt.dataType] = xhr.response
        callback(xhr.status, xhr.statusText, data, xhr.getAllResponseHeaders())
      )
      xhr.open(opt.type, opt.url, yes)
      for key, value of headers
        xhr.setRequestHeader(key, value)
      xhr.responseType = dataType
      xhr.send(opt.data)
    abort: ->
      jqxhr.abort()
  }
)

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
      xhrConfig =
        contentType: 'application/json; charset=utf-8'
        data: dataIn
        dataType: config.dataOut ? 'json'
        processData: no
        type: 'POST'
        headers:
          'X-Key': key
        success: (data, status, xhr) ->
          p.resolve(data)
          reqP.resolve(data)
        error: (err) ->
          log 'Api Error:', err
          if err.status is 401
            login()
            return
          errorStr = "HTTP error: #{err.status}"
          if err.status is 403
            errorStr = err.responseText
          p.reject(errorStr)
          reqP.reject(errorStr)
      $.ajax("#{@base}api/#{path}", xhrConfig)
      return reqP
    xhr()
    return p
  
  makeSSE: (config) ->
    if not window.EventSource then return undefined
    sse = null
    closed = yes
    restartID = null
    onOpen = =>
      # log 'makeSSE onOpen'
      config.open() if config.open
      clearTimeout(restartID) if restartID
    onError = (e) =>
      log 'makeSSE onError', e
      sse.close() if sse
      sse = null
      config.close() if config.close
      restartID = setTimeout(=>
        if sse is null and not closed
          sse = make()
      , 10000)
    onMessage = (e) =>
      # log 'makeSSE onMessage', e
      if e.data is '' then return
      try
        parsed = JSON.parse(e.data)
        config.message(parsed) if config.message
      catch error
        log 'Failed to parse:', e.data, error
      
    make = =>
      evtSource = new EventSource("#{@base}api/link?key=#{@key}")
      evtSource.addEventListener('error', onError)
      evtSource.addEventListener('open', onOpen)
      evtSource.addEventListener('message', onMessage)
      config.connect() if config.connect
      return evtSource
    return {
      on: (event, handler) =>
        return evtSource.addEventListener(event, handler)
      close: () =>
        closed = yes
        sse.close() if sse
        sse = null
        config.close() if config.close
      open: =>
        sse.close() if sse
        closed = no
        sse = make()
      opened: =>
        return sse isnt null
    }

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
  KEY_NET_LINK: 'net_link'
  POLL_INTERVAL_SEC: 60 * 10

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
        networkChangeHandler('Click')
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
    pollingNow = no
    networkChangeHandler = (reason, actual = yes) =>
      # if actual then return
      # log 'networkChangeHandler', navigator.onLine, document[propName]
      if refreshID
        clearTimeout(refreshID)
        refreshID = null
      if (navigator.onLine and document[propName] is no) or forceRefresh
        networkBtn.almostHide(yes)
        # log 'Autorefresh start', reason
        if pollingNow then return
        pollingNow = yes
        return @pollData(actual).always(=>
          pollingNow = no
          # log 'Autorefresh finish'
          refreshID = setTimeout(=>
            networkChangeHandler('Auto-refresh')
          , @POLL_INTERVAL_SEC * 1000)
        )
      else
        networkBtn.almostHide(no)
        return Q(true)
    $(window).on('online', =>
      networkChangeHandler('Online')
    ).on('offline', =>
      networkChangeHandler('Offline')
    )
    $(document).on(eventName, =>
      networkChangeHandler('Visibility')
    )
    setTimeout(=>
      networkChangeHandler('Startup cached', no).then(=>
        networkChangeHandler('Startup actual')
      )
    , 1000)
    sse = @api.makeSSE(
      open: =>
        log 'SSE connected'
        sseBtn.setColor('success')
      close: =>
        log 'SSE closed'
        sseBtn.setColor('failure')
      message: (obj) =>
        # log 'SSE message', obj
        if obj.type is 'sensor'
          # Notify listeners
          @emitDataEvent(obj.data)
      connect: =>
        log 'SSE connect start'
        sseBtn.almostHide(yes)
        sseBtn.setColor()
    )
    if sse
      sseBtn = @makeButton(
        icon: 'wifi'
        target: menuTarget
        handler: =>
          if sse.opened()
            sse.close()
          else
            sse.open()
          sseBtn.almostHide(sse.opened())
      )
      sseBtn.almostHide(no)
      forceLink = @storage.get(@KEY_NET_LINK, 'bool', no)
      if forceLink then sse.open()


  makeUI: (config) ->
    size = $(window)
    # @showError "Size: #{size.width()}x#{size.height()}"
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
    setInterval(() =>
      @redrawSensors()
    , 30*1000)
    @redrawSensors()

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
      if config.prepend
        config.target.prepend(btn)
      else
        config.target.append(btn)
    BTN_CLASSES = ['btn-success', 'btn-failure']
    return {
      text: (text) =>
        btn.find('.text').text(text)
      html: (html) =>
        btn.html(html)
      setColor: (color) =>
        for c in BTN_CLASSES
          btn.removeClass(c)
        btn.addClass("btn-#{color}") if color
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
        for id, detail of details
          if detail.visible
            detail.closeDialog?()
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

  fetchData: (sensors, from, to, forecast = no) ->
    obj =
      series: []
      forecast: forecast
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

  fetchLatest: (sensors, actual = yes) ->
    obj =
      sensors: []
      actual: actual
    for item in sensors
      obj.sensors.push(
        device: item.device
        type: item.type
        index: item.index
        measure: item.measure
      )
    return @api.call('latest',
      body: obj
    ).then((data) =>
      # log 'Data:', data
      for sensor in data.sensors
        @emitDataEvent(sensor)
    , @onError)

  redrawSensors: () ->
    for sensor in @sensors
      sensor.redraw()

  pollData: (actual = yes) ->
    promises = []
    for sensor in @sensors
      p = sensor.refresh(actual)
      if p then promises.push(p)
    return Q.all(promises)

$(document).ready ->
  log 'App started'
  app = new AppController()
  app.loadData()

