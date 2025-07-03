"use strict";

(function() {
    const url = new URL(location);
    url.pathname = "/result";
    history.pushState({}, "", url);
}())