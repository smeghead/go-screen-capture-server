function hello(){
  alert("Hello, world!");
}

window.addEventListener("load", function() { screenshots.init(); }, false);

var screenshots = {
  init: function() {
    var appcontent = document.getElementById("appcontent");
    if(appcontent)
      appcontent.addEventListener("DOMContentLoaded", screenshots.onPageLoad, true);
    dump('addon init\n');
  },

  onPageLoad: function(aEvent) {
    try {
      var doc = aEvent.originalTarget; // doc は "onload" event を起こしたドキュメント
      if (doc.nodeName != "#document") return;
      if (!doc.location.toString().match(/^http/)) {
        dump('not http match: ' + doc.location + '\n');
        return;
      }
      var currentWindow = screenshots.searchWindow(doc);
      if (!currentWindow) return;
      dump('addon currentWindow location: ' + currentWindow.document.location + '\n');

      setTimeout(function() {
        try {
          dump('addon start caputure: ' + currentWindow.document.location + '\n');
          var region = screenshots.getCompletePageRegion(doc);
          var canvas = document.createElementNS("http://www.w3.org/1999/xhtml", "html:canvas");
          screenshots.copyRegionToCanvas(doc, region, canvas);
          dump('addon copied\n');
          var dataUrl = canvas.toDataURL('image/png', '');
          var IOService = Components.Constructor("@mozilla.org/network/io-service;1", "nsIIOService");
          var nsUri = new IOService().newURI(dataUrl, "UTF8", null);

          var nsFile = Components.classes["@mozilla.org/file/local;1"].createInstance(Components.interfaces.nsILocalFile);
          //ディスプレイ番号をプロファイル名から取得する。
          var displayNo = screenshots.getDisplayNo();
          dump('addon displayNo: ' + displayNo + '\n');

          nsFile.initWithPath('/home/smeghead/work/go-screen-capture-server/images/tmp_' + displayNo + '.png');
          dump('addon initWithPath\n');
          //var persist = screenshots.createPersist();
          var persist = Components.classes["@mozilla.org/embedding/browser/nsWebBrowserPersist;1"].createInstance(Components.interfaces.nsIWebBrowserPersist);
          persist.persistFlags = Components.interfaces.nsIWebBrowserPersist.PERSIST_FLAGS_REPLACE_EXISTING_FILES;
          persist.persistFlags |= Components.interfaces.nsIWebBrowserPersist.PERSIST_FLAGS_AUTODETECT_APPLY_CONVERSION;

          dump('addon persist\n');
          persist.saveURI(nsUri, null, null, null, null, nsFile);
          dump('addon saveURI\n');
          //閉じる。
          var gBrowser = window.QueryInterface(Components.interfaces.nsIInterfaceRequestor)
             .getInterface(Components.interfaces.nsIWebNavigation)
             .QueryInterface(Components.interfaces.nsIDocShellTreeItem)
             .rootTreeItem
             .QueryInterface(Components.interfaces.nsIInterfaceRequestor)
             .getInterface(Components.interfaces.nsIDOMWindow).gBrowser;

          var num = gBrowser.browsers.length;
          if (num > 1) {
            for (var i = 0; i < num - 1; i++) {
              var b = gBrowser.getBrowserAtIndex(i);
              try {
                dump(b.currentURI.spec + "\n");
                gBrowser.removeCurrentTab();
              } catch(e) {
                dump(e);
                Components.utils.reportError(e);
              }
            }
          }
        } catch (e) {
          dump('addon ERROR: ' + e);
        }
      }, 2000);
    } catch (e) {
      dump('addon ERROR: ' + e);
    }
  },
  createPersist: function() {
    var persist = Components.classes["@mozilla.org/embedding/browser/nsWebBrowserPersist;1"].createInstance(Components.interfaces.nsIWebBrowserPersist);
    persist.persistFlags = Components.interfaces.nsIWebBrowserPersist.PERSIST_FLAGS_REPLACE_EXISTING_FILES;
    persist.persistFlags |= Components.interfaces.nsIWebBrowserPersist.PERSIST_FLAGS_AUTODETECT_APPLY_CONVERSION;
    return persist;
  },
  getCompletePageRegion : function(doc) {
    var width = screenshots.getDocumentWidth(doc);
    var height = screenshots.getDocumentHeight(doc);
//    if (this.getViewportWidth() > width) width = this.getViewportWidth();
//    if (this.getViewportHeight() > height) height = this.getViewportHeight();
        
    return new screenshots.Region(0, 0, width, height);
  },
  getDocumentHeight : function(doc) {
    if (doc.compatMode == "CSS1Compat") {
      // standards mode
      return doc.documentElement.scrollHeight;
    }
    return doc.body.scrollHeight;
  },
    
  getDocumentWidth : function(doc) {
    if (doc.compatMode == "CSS1Compat") {
      // standards mode
      return doc.documentElement.scrollWidth;
    }
    return doc.body.scrollWidth;
  },
  searchWindow: function(doc) {
    var WindowMediator = Components
            .classes['@mozilla.org/appshell/window-mediator;1']
            .getService(Components.interfaces.nsIWindowMediator);
    var browsers = WindowMediator.getEnumerator('navigator:browser');

    while (browsers.hasMoreElements()) {
      var win = browsers.getNext().gBrowser;
      var num = win.browsers.length;
      for (var i = 0; i < num; i++) {
        var b = win.getBrowserAtIndex(i);
        // アクティブなウィンドウのlocationと同じでない場合は、iframeコンテンツとみなして、キャプチャしない。
        // FIXME この判定条件は不十分(リダイレクトされた場合に、目的のコンテンツであるかどうかを判定できないため)
        if (b.contentWindow.document.location == doc.location) {
          return b.contentWindow;
        }
      }
    }

    return;
  },
  copyRegionToCanvas : function(doc, region, canvas) {
    var context = screenshots.prepareCanvas(canvas, region);
    var currentWindow = screenshots.searchWindow(doc);
    if (!currentWindow) return;
//    dump('copy.');
    context.drawWindow(currentWindow, region.x, region.y, region.width, region.height, '#ffffff');
//    dump('copied.');

    context.restore();
    return context;
  },

  prepareCanvas : function(canvas, region) {
    canvas.width = region.width;
    canvas.height = region.height;
    canvas.style.width = canvas.style.maxwidth = region.width + "px";
    canvas.style.height = canvas.style.maxheight = region.height + "px";
    
    var context = canvas.getContext("2d");
    context.clearRect(region.x, region.y, region.width, region.height);
    context.save();
    return context;
  },
  getCurrentProfileName: function() {
    var path = Cc['@mozilla.org/file/directory_service;1'].getService(Ci.nsIProperties).get('ProfD',Ci.nsILocalFile).path;
    var profileEnum = Cc['@mozilla.org/toolkit/profile-service;1'].createInstance(Ci.nsIToolkitProfileService).profiles;
    while (profileEnum.hasMoreElements()) {
      var profile = profileEnum.getNext().QueryInterface(Ci.nsIToolkitProfile);
      if (profile.rootDir.path == path) {
        return profile.name;
      }
    };
    return '';
  },
  getDisplayNo: function() {
    var profileName = screenshots.getCurrentProfileName();
    dump('profileName: ' + profileName + '\n');
    var results = profileName.match(/P(\d+)/);
    return results[1];
  }

};
screenshots.Region = function(x, y, width, height) {
    this.x = x;
    this.y = y;
    this.width = width;
    this.height = height;
};

// vim: set expandtab sw=2 ts=2 sts=2:
