javascript: (function () {
  if (window.location.hostname !== 'github.com' && window.location.hostname !== 'nxpkg.com') {
    alert('This bookmarklet may only be used on GitHub.com or Nxpkg.com, not ' + window.location.hostname + '.')
    return
  }
  var pats = [
    [
      '^/([^/]+)/([^/]+)/tree/([^/]+)$',
      '/github.com/$1/$2@$3',
      '^/github.com/([^/]+)/([^/@]+)@([^/]+)$',
      '/$1/$2/tree/$3',
    ],
    [
      '^/([^/]+)/([^/]+)/tree/([^/]+)/(.+)$',
      '/github.com/$1/$2@$3/-/tree/$4',
      '^/github.com/([^/]+)/([^/@]+)@([^/]+)/-/tree/(.+)$',
      '/$1/$2/tree/$3/$4',
    ],
    ['^/([^/]+)/([^/]+)/blob/([^/]+)/(.+)$', '/github.com/$1/$2@$3/-/blob/$4', '', ''],
    ['^/([^/]+)/([^/]+)$', '/github.com/$1/$2', '^/github.com/([^/]+)/([^/]+)$', '/$1/$2'],
    ['^/([^/]+)$', '/$1', '^/([^/]+)$', '/$1'],
  ]
  var pathname = window.location.pathname
  if (window.location.hostname === 'nxpkg.com') {
    if (pathname.indexOf('/nxpkg.com/') === 0) {
      pathname = pathname.replace('/nxpkg.com/', '/github.com/')
    } else if (pathname.indexOf('/nxpkg/') === 0) {
      pathname = '/github.com' + pathname
    }
  }

  for (var i = 0; i < pats.length; i++) {
    var pat = pats[i]
    if (window.location.hostname === 'github.com') {
      if (pat[0] === '') {
        continue
      }

      var r = new RegExp(pat[0])
      if (pathname.match(r)) {
        var pathname2 = pathname.replace(r, pat[1])
        window.location = 'https://nxpkg.com' + pathname2
        return
      }
    } else {
      if (pat[2] === '') {
        continue
      }

      var r = new RegExp(pat[2])
      if (pathname.match(r)) {
        var pathname2 = pathname.replace(r, pat[3])
        window.location = 'https://github.com' + pathname2
        return
      }
    }
  }
  alert('Unable to jump to Nxpkg (no matching URL pattern).')
})()
