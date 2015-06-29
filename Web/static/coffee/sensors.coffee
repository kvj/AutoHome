normalize = (a) ->
  n = a.length-1
  if n < 7 then return a
  result = []
  for val, i in a
    v = val
    switch i
      when 0
        v = (39*a[0] + 8*a[1] - 4*(a[2]+a[3]-a[4]) + a[5] - 2*a[6]) / 42
      when 1
        v = (8*a[0] + 19*a[1] + 16*a[2] + 6*a[3] - 4*a[4] - 7*a[5] + 4*a[6]) / 42
      when 2
        v = (-4*a[0] + 16*a[1] + 19*a[2] + 12*a[3] + 2*a[4] - 4*a[5] + 1*a[6]) / 42
      when n-2
        v = (-4*a[n] + 16*a[n-1] + 19*a[n-2] + 12*a[n-3] + 2*a[n-4] - 4*a[n-5] + 1*a[n-6]) / 42
      when n-1
        v = (8*a[n] + 19*a[n-1] + 16*a[n-2] + 6*a[n-3] - 4*a[n-4] - 7*a[n-5] + 4*a[n-6]) / 42
      when n
        v = (39*a[n] + 8*a[n-1] - 4*a[n-2] - 4*a[n-3] + a[n-4] + 4*a[n-5] - 2*a[n-6]) / 42
      else
        v = (7*a[i] + 6*a[i+1] + 6*a[i-1] + 3*a[i+2] + 3*a[i-2] - 2*a[i+3] - 2*a[i-3]) / 21
    result.push(v)
  return result

class ValueSensorDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      text: '...'
      handler: =>
        @room.showDetail(@extra.detail) if @extra.detail
    )
    @app.addDataListener(@config, (data) =>
      @show(data.value)
      # Refresh
    )
    return cont

  refresh: (actual) ->
    return @app.fetchLatest([@config], actual)

  show: (value) ->
    # log 'Value display', @config, value
    text = ''
    if @extra.percent
      text = "#{Math.round(value)}%"
    if @extra.temp
      text = "#{Math.round(value*10) / 10}°"
    @btn.text(text)

registerSensor('sensor_value', ValueSensorDisplay)

class DateSensorDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      text: '...'
    )
    return cont

  redraw: ->
    format = ''
    if @extra.time
      format = 'HH:mm'
    if @extra.date
      format = 'dddd, MMMM d'
    dt = new Date((new Date()).getTime() + 30*1000) # 30 seconds from now
    @btn.text(dt.format(format))

registerSensor('date', DateSensorDisplay)

class DateTimeSensorDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @contents = $("""
    <div class="sensor_dt_root">
      <div class="sensor_dt_time text-large"></div>
      <div class="sensor_dt_date">
        <div class="sensor_dt_day text-large"></div>
        <div class="sensor_dt_week_month">
          <div class="sensor_dt_week text-small"></div>
          <div class="sensor_dt_month text-small"></div>
        </div>
      </div>
    </div>""")
    @btn = @app.makeButton(
      target: cont
      cls: 'no-height'
      contents: @contents
    )
    return cont

  refresh: ->
    format = ''
    dt = new Date()
    @contents.find('.sensor_dt_time').text(dt.format('HH:mm'))
    @contents.find('.sensor_dt_day').text(dt.format('d'))
    @contents.find('.sensor_dt_week').text(dt.format('dddd'))
    @contents.find('.sensor_dt_month').text(dt.format('MMMM'))

registerSensor('date_time', DateTimeSensorDisplay)

class WeatherIconDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      handler: =>
        @room.showDetail(@extra.detail) if @extra.detail
    )
    @app.addDataListener(@config, (data) =>
      @show(data.value)
    )
    return cont

  refresh: (actual) ->
    return @app.fetchLatest([@config], actual)

  show: (value) ->
    icon = 'sun'
    switch value
      when 2 then icon = 'pcloudy' # patly cloudy
      when 3 then icon = 'cloudy' # cloudy
      when 4 then icon = 'rain' # rain
      when 5 then icon = 'lightning' # tstorms
      when 6 then icon = 'snow' # show
      when 7 then icon = 'fog' # fog
    @btn.html("<span class='fa icon-weather icon-#{icon}'></span>")

registerSensor('weather_icon', WeatherIconDisplay)

class LightDisplay extends SensorDisplay

  initialize: (@roomDiv) ->
    @app.addDataListener(@config, (data) =>
      @show(data.value)
    )
    return undefined

  refresh: (actual) ->
    return @app.fetchLatest([@config], actual)

  show: (value) ->
    min = @extra.min
    max = @extra.max
    if value < min
      value = min
    if value > max
      value = max
    intensivity = Math.ceil((max - value) / (max - min) * 255)
    # log 'Light', value, min, max, intensivity
    @roomDiv.css(
      'background-color': "rgb(#{intensivity}, #{intensivity}, #{intensivity})"
    )

registerSensor('light', LightDisplay)

class RoomClassDisplay extends SensorDisplay

  initialize: (@roomDiv) ->
    @app.addDataListener(@config, (data) =>
      @show(data.value)
    )
    @show(0)
    return undefined

  refresh: (actual) ->
    return @app.fetchLatest([@config], actual)

  show: (value) ->
    offCls = @extra.off ? 'off'
    onCls = @extra.on ? 'on'
    if value
      @roomDiv.removeClass("room-cls-#{offCls}").addClass("room-cls-#{onCls}")
    else
      @roomDiv.removeClass("room-cls-#{onCls}").addClass("room-cls-#{offCls}")


registerSensor('class', RoomClassDisplay)

class SensorFlagDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      icon: @extra.icon
    )
    @app.addDataListener(@config, (data) =>
      @show(data.value)
    )
    @show(0)
    return cont

  refresh: (actual) ->
    return @app.fetchLatest([@config], actual)

  show: (value) ->
    # log 'show', @extra
    @btn.almostHide(value == 1)

registerSensor('sensor_flag', SensorFlagDisplay)

class SensorValueDirectionDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @contents = $('<span class="no-wrap"><i class="fa fa-arrow-down"></i><span class="text">...</span></span>')
    @btn = @app.makeButton(
      target: cont
      contents: @contents
    )
    @app.addDataListener(@config, (data) =>
      @value = data.value
    )
    @app.addDataListener(
      device: @config.device
      type: @config.type
      index: @config.index
      measure: @extra.dir
    , (data) =>
      @show(data.value, @value)
    )
    return cont

  refresh: (actual) ->
    return @app.fetchLatest([@config,
      device: @config.device
      type: @config.type
      index: @config.index
      measure: @extra.dir
    ], actual)

  show: (dir, value) ->
    # log 'Wind:', dir, value
    @contents.find('.text').text("#{value}")
    @contents.find('i').css(
      'transform': "rotate(#{dir}deg)"
      '-webkit-transform': "rotate(#{dir}deg)"
    )

registerSensor('direction_value', SensorValueDirectionDisplay)

prepareSeries = (data, doNormalize) ->
    arr = []
    min = 9999
    max = 0
    raw = []
    for item in data
      arr.push([item.ts, item.value])
      raw.push(item.value)
    norm = raw
    normTimes = Math.ceil(raw.length / 50)
    if doNormalize and normTimes > 0
      for j in [0...normTimes]
        norm = normalize(norm)
    for val, j in norm
      if val < min then min = val
      if val > max then max = val
      arr[j][1] = val
    return [arr, min, max]

class InlineGraphDisplay extends SensorDisplay

  COLORS: ['#dc322f', '#268bd2', '#859900']

  initialize: ->

  refresh: ->
    to = new Date().getTime()
    plus = (@extra.hours ? 24) * 60 * 60 * 1000
    if @extra.forecast
      from = to
      to += plus
    else
      from = to - plus
    return @app.fetchData(@config.data, from, to, @extra.forecast).then((data) =>
      series = []
      yaxes = []
      colors = []
      for data, i in data.series
        conf = @config.data[i]
        noNormalize = conf.steps or conf.symbols
        [arr, min, max] = prepareSeries(data, not noNormalize)
        item =
          data: arr
          yaxis: i+1
        series.push(item)
        if conf.fill
          item.lines =
            show: yes
            steps: conf.steps
            fill: yes
        if conf.symbols
          item.points = symbolPoints(arr, conf, WEATHER_DESIGN)
        gap = (max - min) / 2
        if @extra.percent
          gap = 0
          min = 0
          max = 100
        yaxes.push({
          min: min-gap
          max: max+gap
        })
        if conf.color
          col = COLORS[conf.color] ? COLORS.yellow
          colors.push("rgb(#{col[0]}, #{col[1]}, #{col[2]})")
        else
          colors.push(@COLORS[i % @COLORS.length])
      @room.plot(series, colors, yaxes)
    )

registerSensor('inline_graph', InlineGraphDisplay)

COLORS =
  'yellow':  [0xb5, 0x89, 0x00]
  'orange':  [0xcb, 0x4b, 0x16]
  'red':     [0xdc, 0x32, 0x2f]
  'magenta': [0xd3, 0x36, 0x82]
  'violet':  [0x6c, 0x71, 0xc4]
  'blue':    [0x26, 0x8b, 0xd2]
  'cyan':    [0x2a, 0xa1, 0x98]
  'green':   [0x85, 0x99, 0x00]
  'grey':    [0x83, 0x94, 0x96]
  'black':   [0x00, 0x00, 0x00]

color2Color = (name) ->
  col = COLORS[name]
  if not col then return '#000000'
  return "##{col[0].toString(16)}#{col[1].toString(16)}#{col[2].toString(16)}".toUpperCase()

# Generate symbols based on value and definition
symbolPoints = (data, conf, design) ->
  lastX = -1
  lastY = -1
  mapping = {}
  for item, idx in data
    # mapping[item[0]] = (idx % 7) + 1
    mapping[item[0]] = item[1]
    item[1] = 80 # middle line
  spaceSq = (x, y) ->
    return (lastX - x)*(lastX - x) + (lastY - y)*(lastY - y)
  points =
    show: yes
    steps: yes
    radius: 10
    fill: yes
    fillColor: '#FFFFFF'
    symbol: (ctx, x, y, radius, shadow, rawx, rawy) =>
      if shadow then return
      if lastX>=0 and spaceSq(x, y)<2*radius*radius
        return
      else
        lastX = x
        lastY = y
      value = mapping[rawx]
      d = design[value]
      if not value or not d
        return
      ctx.translate(x, y)
      ctx.strokeStyle = color2Color(d.color)
      switch d.shape
        when 'circle'
          mul = 0.5
          ctx.beginPath()
          ctx.arc(0, 0, radius * mul, 0, Math.PI * 2, no)
          ctx.closePath()
        when 'square'
          ctx.rect(-radius * 0.5, -radius * 0.5, radius, radius)
        when 'romb'
          mul = 0.6
          ctx.beginPath()
          ctx.moveTo(-mul * radius, 0)
          ctx.lineTo(0, mul * radius, 0)
          ctx.lineTo(mul * radius, 0)
          ctx.lineTo(0, -mul * radius, 0)
          ctx.closePath()
        when 'triangle2'
          mul = 0.6
          ctx.beginPath()
          ctx.moveTo(-mul * radius, - mul * radius * 0.7)
          ctx.lineTo(mul * radius, - mul* radius * 0.7)
          ctx.lineTo(0, mul * radius * 0.8)
          ctx.closePath()
        when 'triangle'
          mul = 0.6
          ctx.beginPath()
          ctx.moveTo(-mul * radius, mul * radius * 0.7)
          ctx.lineTo(mul * radius, mul* radius * 0.7)
          ctx.lineTo(0, - mul * radius * 0.8)
          ctx.closePath()
      ctx.translate(-x, -y)
  return points

WEATHER_DESIGN =
  1:
    shape: 'circle'
    color: 'orange'
  2:
    shape: 'romb'
    color: 'orange'
  3:
    shape: 'romb'
    color: 'grey'
  4:
    shape: 'triangle2'
    color: 'blue'
  5:
    shape: 'triangle'
    color: 'orange'
  6:
    shape: 'triangle'
    color: 'cyan'
  7:
    shape: 'square'
    color: 'grey'

class DetailGraphDisplay extends SensorDisplay
 
  initialize: ->
    renderOne = (idx) =>
      from = control.from.getTime()
      to = control.to.getTime()
      control.toggleLoading(no)
      @app.fetchData(@config.data[idx].sensors, from, to, @extra.forecast).then((data) =>
        if not control.visible then return
        series = []
        yaxes = []
        colors = []
        yaxes.push({
          min: 0
          max: 100
          show: yes
          position: "right"
          labelWidth: 30
        })
        sources = {}
        for data, i in data.series
          conf = @config.data[idx].sensors[i]
          if conf.source
            map = {}
            for item in data
              map[item.ts] = item.value
            sources[conf.source] = map
            continue
          noNormalize = conf.steps or conf.symbols
          [arr, min, max] = prepareSeries(data, not noNormalize)
          oneItem =
            data: arr
            yaxis: if conf.percent then 1 else yaxes.length+1
          if conf.direction
            do (conf) =>
              lastX = -1
              lastY = -1
              spaceSq = (x, y) ->
                return (lastX - x)*(lastX - x) + (lastY - y)*(lastY - y)
              oneItem.points =
                show: yes
                symbol: (ctx, x, y, radius, shadow, rawx, rawy) =>
                  if shadow then return
                  angle = sources[conf.direction][rawx]
                  if not angle
                    return
                  if lastX>=0 and spaceSq(x, y)<8*radius*radius
                    return
                  else
                    lastX = x
                    lastY = y
                  # log 'Draw', sources, arguments, sources[conf.direction][rawx]
                  ctx.save()
                  ctx.translate(x, y)
                  ctx.rotate(Math.PI*angle / 180 + Math.PI*1/2)
                  ctx.moveTo(-radius / 2, 0)
                  ctx.lineTo(4*radius, 0)
                  ctx.arc(0, 0, radius, 0, Math.PI * 2, no)
                  ctx.restore()
          else if conf.symbols
            oneItem.yaxis = 1
            oneItem.points = symbolPoints(arr, conf, WEATHER_DESIGN)
          else
            oneItem.lines =
              show: yes
              steps: conf.steps
              fill: conf.fill
          series.push(oneItem)
          gap = (max - min) / 2
          if not conf.percent
            yaxes.push({
              labelWidth: 50
              min: if min-gap > 0 then min-gap else 0
              max: max+gap
            })
          col = COLORS[conf.color] ? COLORS.yellow
          colors.push("rgb(#{col[0]}, #{col[1]}, #{col[2]})")
        control.plot(control.divs[idx].div, series, colors, yaxes)
      ).always(->
        control.toggleLoading(yes)
      )
    control = new DetailsDialog(
      app: @app
      forecast: @extra.forecast
      default: @extra.default
      onCreate: (div) =>
        control.divs = []
        for item in @config.data
          sface = control.addSurface($("<div class='surface-item surface-graph'></div>"))
          control.divs.push(sface)
      onRender: =>
        for item, idx in @config.data
          renderOne(idx)

    )
    @room.addDetail(@extra.name, control)

  show: (value) ->

registerSensor('graph', DetailGraphDisplay)

fitImage = (cont, img) ->
  cw = cont.width()
  ch = cont.height()
  iw = img.outerWidth()
  ih = img.outerHeight()
  if iw is 0 or ih is 0
    return no
  if iw > cw or ih > ch
    mul = Math.min(cw / iw, ch / ih)
    img.outerWidth(Math.floor(iw * mul))
    img.outerHeight(Math.floor(ih * mul))
    return yes
  return no

class CameraDisplay extends SensorDisplay

  initialize: ->

    autoRefresh = no

    refresh = (div) =>
      if control.isLoading() or not control.isVisible()
        return
      control.toggleLoading()
      control.scheduleAutoClose()
      return @app.api.call('camera/snapshot',
        body:
          host: @extra.host
          type: @extra.type
        dataOut: 'raw'
      ).then((data) =>
        control.toggleLoading(yes)
        if not control.isVisible() then return # Already hidden
        div.empty()
        dataDiv = $('<div class="surface-item"><img class="surface-img"></div>')
        div.append(dataDiv)
        dataDiv.find('img').on('load', =>
          fitImage(div, dataDiv.find('img'))
        ).attr('src', URL.createObjectURL(data))
        setTimeout(=>
          refresh(div) if autoRefresh
        , 5000) if autoRefresh
      , (error) =>
        control.toggleLoading(yes)
        @app.showError('Failure')
      )

    control = new Dialog(
      app: @app
      autoClose: 60
      onClose: =>
        autoRefresh = no
      onCreate: (div, ndiv) =>
        autoRefresh = no
        @app.makeButton(
          target: ndiv
          prepend: yes
          icon: 'refresh'
          handler: =>
            refresh(div)

        )
        runBtn = @app.makeButton(
          target: ndiv
          prepend: yes
          icon: 'play'
          handler: =>
            autoRefresh = not autoRefresh
            if autoRefresh
              refresh(div)
            runBtn.setColor(if autoRefresh then 'success' else null)
        )
        refresh(div)
    )
    @room.addDetail(@extra.host, control)
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      icon: 'video-camera'
      handler: =>
        @room.showDetail(@extra.host)
    )
    return cont

  refresh: (actual) ->
    return undefined

  show: (value) ->

registerSensor('camera', CameraDisplay)
