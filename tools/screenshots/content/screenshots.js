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
    var doc = aEvent.originalTarget; // doc は "onload" event を起こしたドキュメント
    if (doc.nodeName != "#document") return;
    var currentWindow = screenshots.searchWindow(doc);
    if (!currentWindow) return;
    dump('adon loaded document\n');
    // アクティブなウィンドウのlocationと同じでない場合は、iframeコンテンツとみなして、キャプチャしない。
    if (currentWindow.document.location != doc.location) return;
    dump('addon currentWindow location: ' + currentWindow.document.location + '\n');

    setTimeout(function() {
      try {
        dump('addon start caputure: ' + currentWindow.document.location + '\n');
        var region = screenshots.getCompletePageRegion(doc);
        dump('addon got region: ' + region + '\n');
        var canvas = document.createElementNS("http://www.w3.org/1999/xhtml", "html:canvas");
        dump('addon got canvas: ' + canvas + '\n');
        screenshots.copyRegionToCanvas(doc, region, canvas);
        dump('addon copied\n');
        var dataUrl = canvas.toDataURL('image/png', '');
        dump('addon got dataUrl: ' + '\n');
        var IOService = Components.Constructor("@mozilla.org/network/io-service;1", "nsIIOService");
        dump('addon got IOService: ' + IOService + '\n');
        var nsUri = new IOService().newURI(dataUrl, "UTF8", null);
        dump('addon got nsUri: ' + nsUri + '\n');

        var nsFile = Components.classes["@mozilla.org/file/local;1"].createInstance(Components.interfaces.nsILocalFile);
        //ディスプレイ番号をプロファイル名から取得する。
        var displayNo = screenshots.getDisplayNo();
        dump('addon displayNo: ' + displayNo + '\n');

        nsFile.initWithPath('/home/smeghead/work/go-screen-capture-server/images/tmp_' + displayNo + '.png');
        dump('addon initWithPath\n');
//         var persist = screenshots.createPersist();
        var persist = Components.classes["@mozilla.org/embedding/browser/nsWebBrowserPersist;1"].createInstance(Components.interfaces.nsIWebBrowserPersist);
        persist.persistFlags = Components.interfaces.nsIWebBrowserPersist.PERSIST_FLAGS_REPLACE_EXISTING_FILES;
        persist.persistFlags |= Components.interfaces.nsIWebBrowserPersist.PERSIST_FLAGS_AUTODETECT_APPLY_CONVERSION;

        dump('addon persist\n');
        persist.saveURI(nsUri, null, null, null, null, nsFile); //TODO error handling.
        dump('addon saveURI\n');
      } catch (e) {
        dump("addon " + e);
   //      alert('error: ' + e);
      }
    }, 2000);
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
//        dump(i + ' ' + b.currentURI.spec + '\n');
        if (b.currentURI.spec == doc.location) {
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
