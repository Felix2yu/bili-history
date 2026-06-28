import{ac as Fe,X as wt,o as Ft,N as de,A as R,B,i as A,M as Dt,C as w,D as Y,O as J,T as pe,ad as He,$ as ue,a1 as Be,Z as Ne,E as zt,P as We,Q as je,l as Ut,S as vt,U as Ye,G as qe}from"./CGTz1YfV.js";import{P as Xe}from"./Dw8EzdmS.js";import{getDanmakuFile as Ue,getVideoStream as Ge,getDownloadedVideos as Ze,deleteDownloadedVideo as Ke}from"./BxVAKkzn.js";import{_ as $t}from"./DlAUqK2U.js";import{o as Gt}from"./iPuLo5s-.js";import{n as Zt}from"./DQbmhqKc.js";import{s as Lt}from"./CLIORcfF.js";function Je(n){return n&&n.__esModule&&Object.prototype.hasOwnProperty.call(n,"default")?n.default:n}var bt={exports:{}},Qe=bt.exports,Kt;function tn(){return Kt||(Kt=1,(function(n,t){(function(e,i){n.exports=i()})(Qe,function(){function e(l){return(e=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(d){return typeof d}:function(d){return d&&typeof Symbol=="function"&&d.constructor===Symbol&&d!==Symbol.prototype?"symbol":typeof d})(l)}var i=Object.prototype.toString,o=function(l){if(l===void 0)return"undefined";if(l===null)return"null";var d=e(l);if(d==="boolean")return"boolean";if(d==="string")return"string";if(d==="number")return"number";if(d==="symbol")return"symbol";if(d==="function")return(function(p){return r(p)==="GeneratorFunction"})(l)?"generatorfunction":"function";if((function(p){return Array.isArray?Array.isArray(p):p instanceof Array})(l))return"array";if((function(p){return p.constructor&&typeof p.constructor.isBuffer=="function"?p.constructor.isBuffer(p):!1})(l))return"buffer";if((function(p){try{if(typeof p.length=="number"&&typeof p.callee=="function")return!0}catch(u){if(u.message.indexOf("callee")!==-1)return!0}return!1})(l))return"arguments";if((function(p){return p instanceof Date||typeof p.toDateString=="function"&&typeof p.getDate=="function"&&typeof p.setDate=="function"})(l))return"date";if((function(p){return p instanceof Error||typeof p.message=="string"&&p.constructor&&typeof p.constructor.stackTraceLimit=="number"})(l))return"error";if((function(p){return p instanceof RegExp||typeof p.flags=="string"&&typeof p.ignoreCase=="boolean"&&typeof p.multiline=="boolean"&&typeof p.global=="boolean"})(l))return"regexp";switch(r(l)){case"Symbol":return"symbol";case"Promise":return"promise";case"WeakMap":return"weakmap";case"WeakSet":return"weakset";case"Map":return"map";case"Set":return"set";case"Int8Array":return"int8array";case"Uint8Array":return"uint8array";case"Uint8ClampedArray":return"uint8clampedarray";case"Int16Array":return"int16array";case"Uint16Array":return"uint16array";case"Int32Array":return"int32array";case"Uint32Array":return"uint32array";case"Float32Array":return"float32array";case"Float64Array":return"float64array"}if((function(p){return typeof p.throw=="function"&&typeof p.return=="function"&&typeof p.next=="function"})(l))return"generator";switch(d=i.call(l)){case"[object Object]":return"object";case"[object Map Iterator]":return"mapiterator";case"[object Set Iterator]":return"setiterator";case"[object String Iterator]":return"stringiterator";case"[object Array Iterator]":return"arrayiterator"}return d.slice(8,-1).toLowerCase().replace(/\s/g,"")};function r(l){return l.constructor?l.constructor.name:null}function a(l,d){var p=2<arguments.length&&arguments[2]!==void 0?arguments[2]:["option"];return s(l,d,p),c(l,d,p),(function(u,m,v){var f=o(m),b=o(u);if(f==="object"){if(b!=="object")throw new Error("[Type Error]: '".concat(v.join("."),"' require 'object' type, but got '").concat(b,"'"));Object.keys(m).forEach(function(C){var T=u[C],y=m[C],g=v.slice();g.push(C),s(T,y,g),c(T,y,g),a(T,y,g)})}if(f==="array"){if(b!=="array")throw new Error("[Type Error]: '".concat(v.join("."),"' require 'array' type, but got '").concat(b,"'"));u.forEach(function(C,T){var y=u[T],g=m[T]||m[0],x=v.slice();x.push(T),s(y,g,x),c(y,g,x),a(y,g,x)})}})(l,d,p),l}function s(l,d,p){if(o(d)==="string"){var u=o(l);if(d[0]==="?"&&(d=d.slice(1)+"|undefined"),!(-1<d.indexOf("|")?d.split("|").map(function(m){return m.toLowerCase().trim()}).filter(Boolean).some(function(m){return u===m}):d.toLowerCase().trim()===u))throw new Error("[Type Error]: '".concat(p.join("."),"' require '").concat(d,"' type, but got '").concat(u,"'"))}}function c(l,d,p){if(o(d)==="function"){var u=d(l,o(l),p);if(u!==!0){var m=o(u);throw m==="string"?new Error(u):m==="error"?u:new Error("[Validator Error]: The scheme for '".concat(p.join("."),"' validator require return true, but got '").concat(u,"'"))}}}return a.kindOf=o,a})})(bt)),bt.exports}var en=tn();const pt=Je(en),Ht="5.4.0",ut={properties:["audioTracks","autoplay","buffered","controller","controls","crossOrigin","currentSrc","currentTime","defaultMuted","defaultPlaybackRate","duration","ended","error","loop","mediaGroup","muted","networkState","paused","playbackRate","played","preload","readyState","seekable","seeking","src","startDate","textTracks","videoTracks","volume"],methods:["addTextTrack","canPlayType","load","play","pause"],events:["abort","canplay","canplaythrough","durationchange","emptied","ended","error","loadeddata","loadedmetadata","loadstart","pause","play","playing","progress","ratechange","seeked","seeking","stalled","suspend","timeupdate","volumechange","waiting"],prototypes:["width","height","videoWidth","videoHeight","poster","webkitDecodedFrameCount","webkitDroppedFrameCount","playsInline","webkitSupportsFullscreen","webkitDisplayingFullscreen","onenterpictureinpicture","onleavepictureinpicture","disablePictureInPicture","cancelVideoFrameCallback","requestVideoFrameCallback","getVideoPlaybackQuality","requestPictureInPicture","webkitEnterFullScreen","webkitEnterFullscreen","webkitExitFullScreen","webkitExitFullscreen"]},mt=globalThis?.CUSTOM_USER_AGENT??(typeof navigator<"u"?navigator.userAgent:""),he=/^(?:(?!chrome|android).)*safari/i.test(mt),fe=/iPad|iPhone|iPod/i.test(mt)&&!window.MSStream,me=fe||mt.includes("Macintosh")&&navigator.maxTouchPoints>=1,_=/Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(mt)||me,Tt=typeof window<"u"&&typeof document<"u";function W(n,t=document){return t.querySelector(n)}function gt(n,t=document){return Array.from(t.querySelectorAll(n))}function S(n,t){return n.classList.add(t)}function F(n,t){return n.classList.remove(t)}function Q(n,t){return n.classList.contains(t)}function k(n,t){return t instanceof Element?n.appendChild(t):n.insertAdjacentHTML("beforeend",String(t)),n.lastElementChild||n.lastChild}function Ct(n){return n.parentNode.removeChild(n)}function h(n,t,e){return n.style[t]=e,n}function Mt(n,t){for(const e in t)h(n,e,t[e]);return n}function nn(n,t,e=!0){const i=window.getComputedStyle(n,null).getPropertyValue(t);return e?Number.parseFloat(i):i}function ge(n){return Array.from(n.parentElement.children).filter(t=>t!==n)}function U(n,t){ge(n).forEach(e=>F(e,t)),S(n,t)}function tt(n,t,e="top"){_||(n.setAttribute("aria-label",t),S(n,"hint--rounded"),S(n,`hint--${e}`))}function Bt(n,t=0){const e=n.getBoundingClientRect(),i=window.innerHeight||document.documentElement.clientHeight,o=window.innerWidth||document.documentElement.clientWidth,r=e.top-t<=i&&e.top+e.height+t>=0,a=e.left-t<=o+t&&e.left+e.width+t>=0;return r&&a}function rt(n,t){return Wt(n).includes(t)}function Nt(n,t){return t.parentNode.replaceChild(n,t),n}function D(n){return document.createElement(n)}function ve(n="",t=""){const e=D("i");return S(e,"art-icon"),S(e,`art-icon-${n}`),k(e,t),e}function ye(n,t){let e=document.getElementById(n);e||(e=document.createElement("style"),e.id=n,document.readyState==="loading"?document.addEventListener("DOMContentLoaded",()=>{document.head.appendChild(e)}):(document.head||document.documentElement).appendChild(e)),e.textContent=t}function be(){const n=document.createElement("div");return n.style.display="flex",n.style.display==="flex"}function G(n){return n.getBoundingClientRect()}function xe(n,t){return new Promise((e,i)=>{const o=new Image;o.onload=function(){if(!t||t===1)e(o);else{const r=document.createElement("canvas"),a=r.getContext("2d");r.width=o.width*t,r.height=o.height*t,a.drawImage(o,0,0,r.width,r.height),r.toBlob(s=>{const c=URL.createObjectURL(s),l=new Image;l.onload=function(){e(l)},l.onerror=function(){URL.revokeObjectURL(c),i(new Error(`Image load failed: ${n}`))},l.src=c})}},o.onerror=function(){i(new Error(`Image load failed: ${n}`))},o.src=n})}function Wt(n){if(n.composedPath)return n.composedPath();const t=[];let e=n.target;for(;e;)t.push(e),e=e.parentNode;return!t.includes(window)&&window!==void 0&&t.push(window),t}class we extends Error{constructor(t,e){super(t),typeof Error.captureStackTrace=="function"&&Error.captureStackTrace(this,e||this.constructor),this.name="ArtPlayerError"}}function q(n,t){if(!n)throw new we(t);return n}function ft(n){return n.includes("?")?ft(n.split("?")[0]):n.includes("#")?ft(n.split("#")[0]):n.trim().toLowerCase().split(".").pop()}function ke(n,t){const e=document.createElement("a");e.style.display="none",e.href=n,e.download=t,document.body.appendChild(e),e.click(),document.body.removeChild(e)}function j(n,t,e){return Math.max(Math.min(n,Math.max(t,e)),Math.min(t,e))}function Et(n){return n.charAt(0).toUpperCase()+n.slice(1)}function Z(n){if(!n)return"00:00";const t=r=>r<10?`0${r}`:String(r),e=Math.floor(n/3600),i=Math.floor((n-e*3600)/60),o=Math.floor(n-e*3600-i*60);return(e>0?[e,i,o]:[i,o]).map(t).join(":")}function $e(n){return n.replace(/[&<>'"]/g,t=>({"&":"&amp;","<":"&lt;",">":"&gt;","'":"&#39;",'"':"&quot;"})[t]||t)}function on(n){const t={"&amp;":"&","&lt;":"<","&gt;":">","&#39;":"'","&quot;":'"'},e=new RegExp(`(${Object.keys(t).join("|")})`,"g");return n.replace(e,i=>t[i]||i)}const $=Object.defineProperty,{hasOwnProperty:rn}=Object.prototype;function ht(n,t){return rn.call(n,t)}function Te(n,t){return Object.getOwnPropertyDescriptor(n,t)}function St(...n){const t=e=>e&&typeof e=="object"&&!Array.isArray(e);return n.reduce((e,i)=>(Object.keys(i).forEach(o=>{const r=e[o],a=i[o];Array.isArray(r)&&Array.isArray(a)?e[o]=r.concat(...a):t(r)&&t(a)?e[o]=St(r,a):e[o]=a}),e),{})}function an(n){return n.replace(/(\d\d:\d\d:\d\d)[,.](\d+)/g,(t,e,i)=>{let o=i.slice(0,3);return i.length===1&&(o=`${i}00`),i.length===2&&(o=`${i}0`),`${e},${o}`})}function Ce(n){return`WEBVTT \r
\r
`.concat(an(n).replace(/\{\\([ibu])\}/g,"</$1>").replace(/\{\\([ibu])1\}/g,"<$1>").replace(/\{([ibu])\}/g,"<$1>").replace(/\{\/([ibu])\}/g,"</$1>").replace(/(\d\d:\d\d:\d\d),(\d\d\d)/g,"$1.$2").replace(/\{[\s\S]*?\}/g,"").concat(`\r
\r
`))}function xt(n){return URL.createObjectURL(new Blob([n],{type:"text/vtt"}))}function Me(n){const t=new RegExp("Dialogue:\\s\\d,(\\d+:\\d\\d:\\d\\d.\\d\\d),(\\d+:\\d\\d:\\d\\d.\\d\\d),([^,]*),([^,]*),(?:[^,]*,){4}([\\s\\S]*)$","i");function e(i=""){return i.split(/[:.]/).map((o,r,a)=>{if(r===a.length-1){if(o.length===1)return`.${o}00`;if(o.length===2)return`.${o}0`}else if(o.length===1)return(r===0?"0":":0")+o;return r===0?o:r===a.length-1?`.${o}`:`:${o}`}).join("")}return`WEBVTT

${n.split(/\r?\n/).map(i=>{const o=i.match(t);return o?{start:e(o[1].trim()),end:e(o[2].trim()),text:o[5].replace(/\{[\s\S]*?\}/g,"").replace(/(\\N)/g,`
`).trim().split(/\r?\n/).map(r=>r.trim()).join(`
`)}:null}).filter(i=>i).map((i,o)=>i?`${o+1}
${i.start} --> ${i.end}
${i.text}`:"").filter(i=>i.trim()).join(`

`)}`}function at(n=0){return new Promise(t=>setTimeout(t,n))}function Ee(n,t){let e;return function(...i){const o=()=>(e=null,n.apply(this,i));clearTimeout(e),e=setTimeout(o,t)}}function Se(n,t){let e=!1;return function(...i){e||(n.apply(this,i),e=!0,setTimeout(()=>{e=!1},t))}}const sn=Object.freeze(Object.defineProperty({__proto__:null,ArtPlayerError:we,addClass:S,append:k,assToVtt:Me,capitalize:Et,clamp:j,createElement:D,debounce:Ee,def:$,download:ke,errorHandle:q,escape:$e,get:Te,getComposedPath:Wt,getExt:ft,getIcon:ve,getRect:G,getStyle:nn,has:ht,hasClass:Q,includeFromEvent:rt,inverseClass:U,isBrowser:Tt,isIOS:fe,isIOS13:me,isInViewport:Bt,isMobile:_,isSafari:he,loadImg:xe,mergeDeep:St,query:W,queryAll:gt,remove:Ct,removeClass:F,replaceElement:Nt,secondToTime:Z,setStyle:h,setStyleText:ye,setStyles:Mt,siblings:ge,sleep:at,srtToVtt:Ce,supportsFlex:be,throttle:Se,tooltip:tt,unescape:on,userAgent:mt,vttToBlob:xt},Symbol.toStringTag,{value:"Module"})),Jt="array",P="boolean",H="string",X="number",ot="object",et="function";function ze(n,t,e){return q(t===H||t===X||n instanceof Element,`${e.join(".")} require '${H}' or 'Element' type`)}const ct={html:ze,disable:`?${P}`,name:`?${H}`,index:`?${X}`,style:`?${ot}`,click:`?${et}`,mounted:`?${et}`,tooltip:`?${H}|${X}`,width:`?${X}`,selector:`?${Jt}`,onSelect:`?${et}`,switch:`?${P}`,onSwitch:`?${et}`,range:`?${Jt}`,onRange:`?${et}`,onChange:`?${et}`},Ot={id:H,container:ze,url:H,poster:H,type:H,theme:H,lang:H,volume:X,isLive:P,muted:P,autoplay:P,autoSize:P,autoMini:P,loop:P,flip:P,playbackRate:P,aspectRatio:P,screenshot:P,setting:P,hotkey:P,pip:P,mutex:P,backdrop:P,fullscreen:P,fullscreenWeb:P,subtitleOffset:P,miniProgressBar:P,useSSR:P,playsInline:P,lock:P,gesture:P,fastForward:P,autoPlayback:P,autoOrientation:P,airplay:P,proxy:`?${et}`,plugins:[et],layers:[ct],contextmenu:[ct],settings:[ct],controls:[{...ct,position:(n,t,e)=>{const i=["top","left","right"];return q(i.includes(n),`${e.join(".")} only accept ${i.toString()} as parameters`)}}],quality:[{default:`?${P}`,html:H,url:H}],highlight:[{time:X,text:H}],thumbnails:{url:H,number:X,column:X,width:X,height:X,scale:X},subtitle:{url:H,name:H,type:H,style:ot,escape:P,encoding:H,onVttLoad:et},moreVideoAttr:ot,i18n:ot,icons:ot,cssVar:ot,customType:ot};class it{constructor(t){this.id=0,this.art=t,this.cache=new Map,this.add=this.add.bind(this),this.remove=this.remove.bind(this),this.update=this.update.bind(this)}get show(){return Q(this.art.template.$player,`art-${this.name}-show`)}set show(t){const{$player:e}=this.art.template,i=`art-${this.name}-show`;t?S(e,i):F(e,i),this.art.emit(this.name,t)}toggle(){this.show=!this.show}add(t){const e=typeof t=="function"?t(this.art):t;if(e.html=e.html||"",pt(e,ct),!this.$parent||!this.name||e.disable)return;const i=e.name||`${this.name}${this.id}`;q(!this.cache.has(i),`Can't add an existing [${i}] to the [${this.name}]`),this.id+=1;const o=D("div");S(o,`art-${this.name}`),S(o,`art-${this.name}-${i}`);const r=Array.from(this.$parent.children);o.dataset.index=e.index||this.id;const a=r.find(c=>Number(c.dataset.index)>=Number(o.dataset.index));a?a.insertAdjacentElement("beforebegin",o):k(this.$parent,o),e.html&&k(o,e.html),e.style&&Mt(o,e.style),e.tooltip&&tt(o,e.tooltip);const s=[];if(e.click){const c=this.art.events.proxy(o,"click",l=>{l.preventDefault(),e.click.call(this.art,this,l)});s.push(c)}return e.selector&&["left","right"].includes(e.position)&&this.selector(e,o,s),this[i]=o,this.cache.set(i,{$ref:o,events:s,option:e}),e.mounted&&e.mounted.call(this.art,o),o}remove(t){q(this.cache.has(t),`Can't find [${t}] from the [${this.name}]`);const e=this.cache.get(t);e.option.beforeUnmount&&e.option.beforeUnmount.call(this.art,e.$ref);for(const i of e.events)this.art.events.remove(i);this.cache.delete(t),delete this[t],Ct(e.$ref)}update(t){if(this.cache.has(t.name)){const e=this.cache.get(t.name);t=Object.assign(e.option,t),this.remove(t.name)}return this.add(t)}}function ln(n){return t=>{const{i18n:e,constructor:{ASPECT_RATIO:i}}=t,o=i.map(r=>`<span data-value="${r}">${r==="default"?e.get("Default"):r}</span>`).join("");return{...n,html:`${e.get("Aspect Ratio")}: ${o}`,click:(r,a)=>{const{value:s}=a.target.dataset;s&&(t.aspectRatio=s,r.show=!1)},mounted:r=>{const a=W('[data-value="default"]',r);a&&U(a,"art-current"),t.on("aspectRatio",s=>{const c=gt("span",r).find(l=>l.dataset.value===s);c&&U(c,"art-current")})}}}}function cn(n){return t=>({...n,html:t.i18n.get("Close"),click:e=>{e.show=!1}})}function dn(n){return t=>{const{i18n:e,constructor:{FLIP:i}}=t,o=i.map(r=>`<span data-value="${r}">${e.get(Et(r))}</span>`).join("");return{...n,html:`${e.get("Video Flip")}: ${o}`,click:(r,a)=>{const{value:s}=a.target.dataset;s&&(t.flip=s.toLowerCase(),r.show=!1)},mounted:r=>{const a=W('[data-value="normal"]',r);a&&U(a,"art-current"),t.on("flip",s=>{const c=gt("span",r).find(l=>l.dataset.value===s);c&&U(c,"art-current")})}}}}function pn(n){return t=>({...n,html:t.i18n.get("Video Info"),click:e=>{t.info.show=!0,e.show=!1}})}function un(n){return t=>{const{i18n:e,constructor:{PLAYBACK_RATE:i}}=t,o=i.map(r=>`<span data-value="${r}">${r===1?e.get("Normal"):r.toFixed(1)}</span>`).join("");return{...n,html:`${e.get("Play Speed")}: ${o}`,click:(r,a)=>{const{value:s}=a.target.dataset;s&&(t.playbackRate=Number(s),r.show=!1)},mounted:r=>{const a=W('[data-value="1"]',r);a&&U(a,"art-current"),t.on("video:ratechange",()=>{const s=gt("span",r).find(c=>Number(c.dataset.value)===t.playbackRate);s&&U(s,"art-current")})}}}}function hn(n){const t=Tt?location.href:"";return{...n,html:`<a href="https://artplayer.org?ref=${encodeURIComponent(t)}" target="_blank" style="width:100%;">ArtPlayer ${Ht}</a>`}}class fn extends it{constructor(t){super(t),this.name="contextmenu",this.$parent=t.template.$contextmenu,_||this.init()}init(){const{option:t,proxy:e,template:{$player:i,$contextmenu:o}}=this.art;t.playbackRate&&this.add(un({name:"playbackRate",index:10})),t.aspectRatio&&this.add(ln({name:"aspectRatio",index:20})),t.flip&&this.add(dn({name:"flip",index:30})),this.add(pn({name:"info",index:40})),this.add(hn({name:"version",index:50})),this.add(cn({name:"close",index:60}));for(let r=0;r<t.contextmenu.length;r++)this.add(t.contextmenu[r]);e(i,"contextmenu",r=>{if(!this.art.constructor.CONTEXTMENU)return;r.preventDefault(),this.show=!0;const a=r.clientX,s=r.clientY,{height:c,width:l,left:d,top:p}=G(i),{height:u,width:m}=G(o);let v=a-d,f=s-p;a+m>d+l&&(v=l-m),s+u>p+c&&(f=c-u),Mt(o,{top:`${f}px`,left:`${v}px`})}),e(i,"click",r=>{rt(r,o)||(this.show=!1)}),this.art.on("blur",()=>{this.show=!1})}}function mn(n){return t=>({...n,tooltip:t.i18n.get("AirPlay"),mounted:e=>{const{proxy:i,icons:o}=t;k(e,o.airplay),i(e,"click",()=>t.airplay())}})}function gn(n){return t=>({...n,tooltip:t.i18n.get("Fullscreen"),mounted:e=>{const{proxy:i,icons:o,i18n:r}=t,a=k(e,o.fullscreenOn),s=k(e,o.fullscreenOff);h(s,"display","none"),i(e,"click",()=>{t.fullscreen=!t.fullscreen}),t.on("fullscreen",c=>{c?(tt(e,r.get("Exit Fullscreen")),h(a,"display","none"),h(s,"display","inline-flex")):(tt(e,r.get("Fullscreen")),h(a,"display","inline-flex"),h(s,"display","none"))})}})}function vn(n){return t=>({...n,tooltip:t.i18n.get("Web Fullscreen"),mounted:e=>{const{proxy:i,icons:o,i18n:r}=t,a=k(e,o.fullscreenWebOn),s=k(e,o.fullscreenWebOff);h(s,"display","none"),i(e,"click",()=>{t.fullscreenWeb=!t.fullscreenWeb}),t.on("fullscreenWeb",c=>{c?(tt(e,r.get("Exit Web Fullscreen")),h(a,"display","none"),h(s,"display","inline-flex")):(tt(e,r.get("Web Fullscreen")),h(a,"display","inline-flex"),h(s,"display","none"))})}})}function yn(n){return t=>({...n,tooltip:t.i18n.get("PIP Mode"),mounted:e=>{const{proxy:i,icons:o,i18n:r}=t;k(e,o.pip),i(e,"click",()=>{t.pip=!t.pip}),t.on("pip",a=>{tt(e,r.get(a?"Exit PIP Mode":"PIP Mode"))})}})}function bn(n){return t=>({...n,mounted:e=>{const{proxy:i,icons:o,i18n:r}=t,a=k(e,o.play),s=k(e,o.pause);tt(a,r.get("Play")),tt(s,r.get("Pause")),i(a,"click",()=>{t.play()}),i(s,"click",()=>{t.pause()});function c(){h(a,"display","flex"),h(s,"display","none")}function l(){h(a,"display","none"),h(s,"display","flex")}t.playing?l():c(),t.on("video:playing",()=>{l()}),t.on("video:pause",()=>{c()})}})}function dt(n,t){const{$progress:e}=n.template,{left:i}=G(e),o=_?t.touches[0].clientX:t.clientX,r=j(o-i,0,e.clientWidth),a=r/e.clientWidth*n.duration,s=Z(a),c=j(r/e.clientWidth,0,1);return{second:a,time:s,width:r,percentage:c}}function Le(n,t){if(n.isRotate){const e=t.touches[0].clientY/n.height,i=e*n.duration;n.emit("setBar","played",e,t),n.seek=i}else{const{second:e,percentage:i}=dt(n,t);n.emit("setBar","played",i,t),n.seek=e}}function xn(n){return t=>{const{icons:e,option:i,proxy:o}=t,{$player:r,$progress:a}=t.template;return{...n,html:`
                <div class="art-control-progress-inner">
                    <div class="art-progress-hover"></div>
                    <div class="art-progress-loaded"></div>
                    <div class="art-progress-played"></div>
                    <div class="art-progress-highlight"></div>
                    <div class="art-progress-indicator"></div>
                    <div class="art-progress-tip">00:00</div>
                </div>
            `,mounted:s=>{let c=null,l=!1;const d=W(".art-progress-hover",s),p=W(".art-progress-loaded",s),u=W(".art-progress-played",s),m=W(".art-progress-highlight",s),v=W(".art-progress-indicator",s),f=W(".art-progress-tip",s);e.indicator?k(v,e.indicator):h(v,"backgroundColor","var(--art-theme)");function b(g){const{width:x}=dt(t,g),{text:M}=g.target.dataset;f.textContent=M;const E=f.clientWidth;x<=E/2?h(f,"left",0):x>s.clientWidth-E/2?h(f,"left",`${s.clientWidth-E}px`):h(f,"left",`${x-E/2}px`)}function C(g,x){const{width:M,time:E}=x||dt(t,g);f.textContent=E||"00:00";const L=f.clientWidth;M<=L/2?h(f,"left",0):M>s.clientWidth-L/2?h(f,"left",`${s.clientWidth-L}px`):h(f,"left",`${M-L/2}px`)}function T(){m.textContent="";for(let g=0;g<i.highlight.length;g++){const x=i.highlight[g],M=j(x.time,0,t.duration)/t.duration*100,E=`<span data-text="${x.text}" data-time="${x.time}" style="left: ${M}%"></span>`;k(m,E)}}function y(g,x,M){const E=g==="played"&&M&&_;if(g==="loaded"&&h(p,"width",`${x*100}%`),g==="hover"&&(h(d,"width",`${x*100}%`),rt(M,m)?b(M):C(M),x===0?F(r,"art-progress-hover"):S(r,"art-progress-hover")),g==="played"&&(h(u,"width",`${x*100}%`),h(v,"left",`${x*100}%`)),E){S(r,"art-progress-hover");const L=s.clientWidth*x,I=Z(x*t.duration);C(M,{width:L,time:I}),clearTimeout(c),c=setTimeout(()=>{F(r,"art-progress-hover")},500)}}t.on("setBar",y),t.on("video:loadedmetadata",T),t.constructor.USE_RAF?t.on("raf",()=>{t.emit("setBar","played",t.played),t.emit("setBar","loaded",t.loaded)}):(t.on("video:timeupdate",()=>{t.emit("setBar","played",t.played)}),t.on("video:progress",()=>{t.emit("setBar","loaded",t.loaded)}),t.on("video:ended",()=>{t.emit("setBar","played",1)})),t.emit("setBar","loaded",t.loaded||0),_||(o(a,"click",g=>{g.target!==v&&Le(t,g)}),o(a,"mousemove",g=>{const{percentage:x}=dt(t,g);t.emit("setBar","hover",x,g)}),o(a,"mouseleave",g=>{t.emit("setBar","hover",0,g)}),o(a,"mousedown",g=>{l=g.button===0}),t.on("document:mousemove",g=>{if(l){const{second:x,percentage:M}=dt(t,g);t.emit("setBar","played",M,g),t.seek=x}}),t.on("document:mouseup",()=>{l&&(l=!1)}))}}}}function wn(n){return t=>({...n,tooltip:t.i18n.get("Screenshot"),mounted:e=>{const{proxy:i,icons:o}=t;k(e,o.screenshot),i(e,"click",()=>{t.screenshot()})}})}function kn(n){return t=>({...n,tooltip:t.i18n.get("Show Setting"),mounted:e=>{const{proxy:i,icons:o,i18n:r}=t;k(e,o.setting),i(e,"click",()=>{t.setting.toggle(),t.setting.resize()}),t.on("setting",a=>{tt(e,r.get(a?"Hide Setting":"Show Setting"))})}})}function $n(n){return t=>({...n,style:_?{fontSize:"12px",padding:"0 5px"}:{cursor:"auto",padding:"0 10px"},mounted:e=>{function i(){const r=`${Z(t.currentTime)} / ${Z(t.duration)}`;r!==e.textContent&&(e.textContent=r)}i();const o=["video:loadedmetadata","video:timeupdate","video:progress"];for(let r=0;r<o.length;r++)t.on(o[r],i)}})}function Tn(n){return t=>({...n,mounted:e=>{const{proxy:i,icons:o}=t,r=k(e,o.volume),a=k(e,o.volumeClose),s=k(e,'<div class="art-volume-panel"></div>'),c=k(s,'<div class="art-volume-inner"></div>'),l=k(c,'<div class="art-volume-val"></div>'),d=k(c,'<div class="art-volume-slider"></div>'),p=k(d,'<div class="art-volume-handle"></div>'),u=k(p,'<div class="art-volume-loaded"></div>'),m=k(d,'<div class="art-volume-indicator"></div>');function v(b){const{top:C,height:T}=G(d);return 1-(b.clientY-C)/T}function f(){if(t.muted||t.volume===0)h(r,"display","none"),h(a,"display","flex"),h(m,"top","100%"),h(u,"top","100%"),l.textContent=0;else{const b=t.volume*100;h(r,"display","flex"),h(a,"display","none"),h(m,"top",`${100-b}%`),h(u,"top",`${100-b}%`),l.textContent=Math.floor(b)}}if(f(),t.on("video:volumechange",f),i(r,"click",()=>{t.muted=!0}),i(a,"click",()=>{t.muted=!1}),_)h(s,"display","none");else{let b=!1;i(d,"mousedown",C=>{b=C.button===0,t.volume=v(C)}),t.on("document:mousemove",C=>{b&&(t.muted=!1,t.volume=v(C))}),t.on("document:mouseup",()=>{b&&(b=!1)})}}})}class Cn extends it{constructor(t){super(t),this.isHover=!1,this.name="control",this.timer=Date.now();const{constructor:e}=t,{$player:i,$bottom:o}=this.art.template;t.on("mousemove",()=>{_||(this.show=!0)}),t.on("click",()=>{_?this.toggle():this.show=!0}),t.on("document:mousemove",r=>{this.isHover=rt(r,o)}),t.on("video:timeupdate",()=>{!t.setting.show&&!this.isHover&&!t.isInput&&t.playing&&this.show&&Date.now()-this.timer>=e.CONTROL_HIDE_TIME&&(this.show=!1)}),t.on("control",r=>{r?(F(i,"art-hide-cursor"),S(i,"art-hover"),this.timer=Date.now()):(S(i,"art-hide-cursor"),F(i,"art-hover"))}),this.init()}init(){const{option:t}=this.art;t.isLive||this.add(xn({name:"progress",position:"top",index:10})),this.add({name:"thumbnails",position:"top",index:20}),this.add(bn({name:"playAndPause",position:"left",index:10})),this.add(Tn({name:"volume",position:"left",index:20})),t.isLive||this.add($n({name:"time",position:"left",index:30})),t.quality.length&&at().then(()=>{this.art.quality=t.quality}),t.screenshot&&!_&&this.add(wn({name:"screenshot",position:"right",index:20})),t.setting&&this.add(kn({name:"setting",position:"right",index:30})),t.pip&&this.add(yn({name:"pip",position:"right",index:40})),t.airplay&&window.WebKitPlaybackTargetAvailabilityEvent&&this.add(mn({name:"airplay",position:"right",index:50})),t.fullscreenWeb&&this.add(vn({name:"fullscreenWeb",position:"right",index:60})),t.fullscreen&&this.add(gn({name:"fullscreen",position:"right",index:70}));for(let e=0;e<t.controls.length;e++)this.add(t.controls[e])}add(t){const e=typeof t=="function"?t(this.art):t,{$progress:i,$controlsLeft:o,$controlsRight:r}=this.art.template;switch(e.position){case"top":this.$parent=i;break;case"left":this.$parent=o;break;case"right":this.$parent=r;break;default:q(!1,"Control option.position must one of 'top', 'left', 'right'");break}super.add(e)}check(t){if(t){t.$control_value.innerHTML=t.html;for(let e=0;e<t.$control_option.length;e++){const i=t.$control_option[e];i.default=i===t,i.default&&U(i.$control_item,"art-current")}}}selector(t,e,i){const{proxy:o}=this.art.events;S(e,"art-control-selector");const r=D("div");S(r,"art-selector-value"),k(r,t.html),e.textContent="",k(e,r);const a=D("div");S(a,"art-selector-list"),k(e,a);for(let c=0;c<t.selector.length;c++){const l=t.selector[c],d=D("div");S(d,"art-selector-item"),l.default&&S(d,"art-current"),d.dataset.index=c,d.dataset.value=l.value,d.innerHTML=l.html,k(a,d),$(l,"$control_option",{get:()=>t.selector}),$(l,"$control_item",{get:()=>d}),$(l,"$control_value",{get:()=>r})}const s=o(a,"click",async c=>{const l=Wt(c),d=t.selector.find(p=>p.$control_item===l.find(u=>p.$control_item===u));this.check(d),t.onSelect&&(r.innerHTML=await t.onSelect.call(this.art,d,d.$control_item,c))});i.push(s)}}function Mn(n,t){const{constructor:e,template:{$player:i,$video:o}}=n;function r(s){rt(s,i)?(n.isInput=s.target.tagName==="INPUT",n.isFocus=!0,n.emit("focus",s)):(n.isInput=!1,n.isFocus=!1,n.emit("blur",s))}n.on("document:click",r),n.on("document:contextmenu",r);let a=[];t.proxy(o,"click",s=>{const c=Date.now();a.push(c);const{MOBILE_CLICK_PLAY:l,DBCLICK_TIME:d,MOBILE_DBCLICK_PLAY:p,DBCLICK_FULLSCREEN:u}=e,m=a.filter(v=>c-v<=d);switch(m.length){case 1:n.emit("click",s),_?!n.isLock&&l&&n.toggle():n.toggle(),a=m;break;case 2:n.emit("dblclick",s),_?!n.isLock&&p&&n.toggle():u&&(n.fullscreen=!n.fullscreen),a=[];break;default:a=[]}})}function En(n,t){return Math.atan2(t,n)*180/Math.PI}function Sn(n,t,e,i){const o=t-i,r=e-n;let a=0;if(Math.abs(r)<2&&Math.abs(o)<2)return a;const s=En(r,o);return s>=-45&&s<45?a=4:s>=45&&s<135?a=1:s>=-135&&s<-45?a=2:(s>=135&&s<=180||s>=-180&&s<-135)&&(a=3),a}function zn(n,t){if(_&&!n.option.isLive){const{$video:e,$progress:i}=n.template;let o=null,r=!1,a=0,s=0,c=0;const l=u=>{if(u.touches.length===1&&!n.isLock){o===i&&Le(n,u),r=!0;const{pageX:m,pageY:v}=u.touches[0];a=m,s=v,c=n.currentTime}},d=u=>{if(u.touches.length===1&&r&&n.duration){const{pageX:m,pageY:v}=u.touches[0],f=Sn(a,s,m,v),b=[3,4].includes(f),C=[1,2].includes(f);if(b&&!n.isRotate||C&&n.isRotate){const y=j((m-a)/n.width,-1,1),g=j((v-s)/n.height,-1,1),x=n.isRotate?g:y,M=o===e?n.constructor.TOUCH_MOVE_RATIO:1,E=j(c+n.duration*x*M,0,n.duration);n.seek=E,n.emit("setBar","played",j(E/n.duration,0,1),u),n.notice.show=`${Z(E)} / ${Z(n.duration)}`}}},p=()=>{r&&(a=0,s=0,c=0,r=!1,o=null)};n.option.gesture&&(t.proxy(e,"touchstart",u=>{o=e,l(u)}),t.proxy(e,"touchmove",d)),t.proxy(i,"touchstart",u=>{o=i,l(u)}),t.proxy(i,"touchmove",d),n.on("document:touchend",p)}}function Ln(n,t){const e=["click","mouseup","keydown","touchend","touchmove","mousemove","pointerup","contextmenu","pointermove","visibilitychange","webkitfullscreenchange"],i=["resize","scroll","orientationchange"],o=[];function r(a={}){for(let c=0;c<o.length;c++)t.remove(o[c]);o.length=0;const{$player:s}=n.template;e.forEach(c=>{const l=a.document||s.ownerDocument||document,d=t.proxy(l,c,p=>{n.emit(`document:${c}`,p)});o.push(d)}),i.forEach(c=>{const l=a.window||s.ownerDocument?.defaultView||window,d=t.proxy(l,c,p=>{n.emit(`window:${c}`,p)});o.push(d)})}r(),t.bindGlobalEvents=r}function Pn(n,t){const{$player:e}=n.template;t.hover(e,i=>{S(e,"art-hover"),n.emit("hover",!0,i)},i=>{F(e,"art-hover"),n.emit("hover",!1,i)})}function _n(n,t){const{$player:e}=n.template;t.proxy(e,"mousemove",i=>{n.emit("mousemove",i)})}function In(n,t){const{option:e,constructor:i}=n;n.on("resize",()=>{const{aspectRatio:r,notice:a}=n;n.state==="standard"&&e.autoSize&&n.autoSize(),n.aspectRatio=r,a.show=""});const o=Ee(()=>n.emit("resize"),i.RESIZE_TIME);n.on("window:orientationchange",()=>o()),n.on("window:resize",()=>o()),screen&&screen.orientation&&screen.orientation.onchange&&t.proxy(screen.orientation,"change",()=>o())}function An(n){if(n.constructor.USE_RAF){let t=null;(function e(){n.playing&&n.emit("raf"),n.isDestroy||(t=requestAnimationFrame(e))})(),n.on("destroy",()=>{cancelAnimationFrame(t)})}}function Rn(n){const{option:t,constructor:e,template:{$container:i}}=n,o=Se(()=>{n.emit("view",Bt(i,e.SCROLL_GAP))},e.SCROLL_TIME);n.on("window:scroll",()=>o()),n.on("view",r=>{t.autoMini&&(n.mini=!r)})}class Vn{constructor(t){this.destroyEvents=new Set,this.proxy=this.proxy.bind(this),this.hover=this.hover.bind(this),Mn(t,this),Pn(t,this),_n(t,this),In(t,this),zn(t,this),Rn(t),Ln(t,this),An(t)}proxy(t,e,i,o={}){if(Array.isArray(e))return e.map(a=>this.proxy(t,a,i,o));t.addEventListener(e,i,o);const r=()=>t.removeEventListener(e,i,o);return this.destroyEvents.add(r),r}hover(t,e,i){e&&this.proxy(t,"mouseenter",e),i&&this.proxy(t,"mouseleave",i)}remove(t){if(this.destroyEvents.has(t))try{t()}catch(e){console.warn("Failed to remove event listener:",e)}finally{this.destroyEvents.delete(t)}}destroy(){for(const t of this.destroyEvents)try{t()}catch(e){console.warn("Failed to destroy event listener:",e)}this.destroyEvents.clear()}}class Dn{constructor(t){this.art=t,this.keys={},_||this.init()}init(){const{constructor:t}=this.art;this.art.option.hotkey&&(this.add("Escape",()=>{this.art.fullscreenWeb&&(this.art.fullscreenWeb=!1)}),this.add("Space",()=>{this.art.toggle()}),this.add("ArrowLeft",()=>{this.art.backward=t.SEEK_STEP}),this.add("ArrowUp",()=>{this.art.volume+=t.VOLUME_STEP}),this.add("ArrowRight",()=>{this.art.forward=t.SEEK_STEP}),this.add("ArrowDown",()=>{this.art.volume-=t.VOLUME_STEP})),this.art.on("document:keydown",e=>{if(this.art.isFocus){const i=document.activeElement.tagName.toUpperCase(),o=document.activeElement.getAttribute("contenteditable");if(i!=="INPUT"&&i!=="TEXTAREA"&&o!==""&&o!=="true"&&!e.altKey&&!e.ctrlKey&&!e.metaKey&&!e.shiftKey){const r=this.keys[e.code];if(r){e.preventDefault();for(let a=0;a<r.length;a++)r[a].call(this.art,e);this.art.emit("hotkey",e)}}}this.art.emit("keydown",e)})}add(t,e){return this.keys[t]?this.keys[t].includes(e)||this.keys[t].push(e):this.keys[t]=[e],this}remove(t,e){if(this.keys[t]){const i=this.keys[t].indexOf(e);i!==-1&&this.keys[t].splice(i,1),this.keys[t].length===0&&delete this.keys[t]}return this}}const Pe={"Video Info":"统计信息",Close:"关闭","Video Load Failed":"加载失败",Volume:"音量",Play:"播放",Pause:"暂停",Rate:"速度",Mute:"静音","Video Flip":"画面翻转",Horizontal:"水平",Vertical:"垂直",Reconnect:"重新连接","Show Setting":"显示设置","Hide Setting":"隐藏设置",Screenshot:"截图","Play Speed":"播放速度","Aspect Ratio":"画面比例",Default:"默认",Normal:"正常",Open:"打开","Switch Video":"切换","Switch Subtitle":"切换字幕",Fullscreen:"全屏","Exit Fullscreen":"退出全屏","Web Fullscreen":"网页全屏","Exit Web Fullscreen":"退出网页全屏","Mini Player":"迷你播放器","PIP Mode":"开启画中画","Exit PIP Mode":"退出画中画","PIP Not Supported":"不支持画中画","Fullscreen Not Supported":"不支持全屏","Subtitle Offset":"字幕偏移","Last Seen":"上次看到","Jump Play":"跳转播放",AirPlay:"隔空播放","AirPlay Not Available":"隔空播放不可用"};typeof window<"u"&&(window["artplayer-i18n-zh-cn"]=Pe);class On{constructor(t){this.art=t,this.languages={"zh-cn":Pe},this.language={},this.update(t.option.i18n)}init(){const t=this.art.option.lang.toLowerCase();this.language=this.languages[t]||{}}get(t){return this.language[t]||t}update(t){this.languages=St(this.languages,t),this.init()}}const Fn=`<svg width="18px" height="18px" viewBox="0 0 18 18" xmlns="http://www.w3.org/2000/svg">
    <g>
        <path d="M16,1 L2,1 C1.447,1 1,1.447 1,2 L1,12 C1,12.553 1.447,13 2,13 L5,13 L5,11 L3,11 L3,3 L15,3 L15,11 L13,11 L13,13 L16,13 C16.553,13 17,12.553 17,12 L17,2 C17,1.447 16.553,1 16,1 L16,1 Z"></path>
        <polygon points="4 17 14 17 9 11"></polygon>
    </g>
</svg>
`,Hn=`<svg xmlns="http://www.w3.org/2000/svg" height="32" width="32" version="1.1" viewBox="0 0 32 32">
    <path d="M 19.41,20.09 14.83,15.5 19.41,10.91 18,9.5 l -6,6 6,6 z" />
</svg>`,Bn=`<svg xmlns="http://www.w3.org/2000/svg" height="32" width="32" version="1.1" viewBox="0 0 32 32">
    <path d="m 12.59,20.34 4.58,-4.59 -4.58,-4.59 1.41,-1.41 6,6 -6,6 z" />
</svg>`,Nn='<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 88 88" preserveAspectRatio="xMidYMid meet" style="width: 100%; height: 100%; transform: translate3d(0px, 0px, 0px);"><defs><clipPath id="__lottie_element_216"><rect width="88" height="88" x="0" y="0"></rect></clipPath></defs><g clip-path="url(#__lottie_element_216)"><g transform="matrix(1,0,0,1,44,44)" opacity="1" style="display: block;"><g opacity="1" transform="matrix(1,0,0,1,0,0)"><path fill-opacity="1" d=" M12.437999725341797,-12.70199966430664 C12.437999725341797,-12.70199966430664 9.618000030517578,-9.881999969482422 9.618000030517578,-9.881999969482422 C8.82800006866455,-9.092000007629395 8.82800006866455,-7.831999778747559 9.618000030517578,-7.052000045776367 C9.618000030517578,-7.052000045776367 16.687999725341797,0.017999999225139618 16.687999725341797,0.017999999225139618 C16.687999725341797,0.017999999225139618 9.618000030517578,7.0879998207092285 9.618000030517578,7.0879998207092285 C8.82800006866455,7.877999782562256 8.82800006866455,9.137999534606934 9.618000030517578,9.918000221252441 C9.618000030517578,9.918000221252441 12.437999725341797,12.748000144958496 12.437999725341797,12.748000144958496 C13.227999687194824,13.527999877929688 14.48799991607666,13.527999877929688 15.267999649047852,12.748000144958496 C15.267999649047852,12.748000144958496 26.58799934387207,1.437999963760376 26.58799934387207,1.437999963760376 C27.368000030517578,0.6579999923706055 27.368000030517578,-0.6119999885559082 26.58799934387207,-1.3919999599456787 C26.58799934387207,-1.3919999599456787 15.267999649047852,-12.70199966430664 15.267999649047852,-12.70199966430664 C14.48799991607666,-13.491999626159668 13.227999687194824,-13.491999626159668 12.437999725341797,-12.70199966430664z M-12.442000389099121,-12.70199966430664 C-13.182000160217285,-13.442000389099121 -14.362000465393066,-13.482000350952148 -15.142000198364258,-12.821999549865723 C-15.142000198364258,-12.821999549865723 -15.272000312805176,-12.70199966430664 -15.272000312805176,-12.70199966430664 C-15.272000312805176,-12.70199966430664 -26.582000732421875,-1.3919999599456787 -26.582000732421875,-1.3919999599456787 C-27.32200050354004,-0.6520000100135803 -27.36199951171875,0.5180000066757202 -26.70199966430664,1.3079999685287476 C-26.70199966430664,1.3079999685287476 -26.582000732421875,1.437999963760376 -26.582000732421875,1.437999963760376 C-26.582000732421875,1.437999963760376 -15.272000312805176,12.748000144958496 -15.272000312805176,12.748000144958496 C-14.531999588012695,13.48799991607666 -13.362000465393066,13.527999877929688 -12.571999549865723,12.868000030517578 C-12.571999549865723,12.868000030517578 -12.442000389099121,12.748000144958496 -12.442000389099121,12.748000144958496 C-12.442000389099121,12.748000144958496 -9.612000465393066,9.918000221252441 -9.612000465393066,9.918000221252441 C-8.871999740600586,9.178000450134277 -8.831999778747559,8.008000373840332 -9.501999855041504,7.2179999351501465 C-9.501999855041504,7.2179999351501465 -9.612000465393066,7.0879998207092285 -9.612000465393066,7.0879998207092285 C-9.612000465393066,7.0879998207092285 -16.68199920654297,0.017999999225139618 -16.68199920654297,0.017999999225139618 C-16.68199920654297,0.017999999225139618 -9.612000465393066,-7.052000045776367 -9.612000465393066,-7.052000045776367 C-8.871999740600586,-7.791999816894531 -8.831999778747559,-8.961999893188477 -9.501999855041504,-9.751999855041504 C-9.501999855041504,-9.751999855041504 -9.612000465393066,-9.881999969482422 -9.612000465393066,-9.881999969482422 C-9.612000465393066,-9.881999969482422 -12.442000389099121,-12.70199966430664 -12.442000389099121,-12.70199966430664z M28,-28 C32.41999816894531,-28 36,-24.420000076293945 36,-20 C36,-20 36,20 36,20 C36,24.420000076293945 32.41999816894531,28 28,28 C28,28 -28,28 -28,28 C-32.41999816894531,28 -36,24.420000076293945 -36,20 C-36,20 -36,-20 -36,-20 C-36,-24.420000076293945 -32.41999816894531,-28 -28,-28 C-28,-28 28,-28 28,-28z" data-darkreader-inline-fill="" style="--darkreader-inline-fill:#a8a6a4;"></path></g></g></g></svg>',Wn=`<svg xmlns="http://www.w3.org/2000/svg" version="1.1" viewBox="0 0 24 24" style="width: 100%; height: 100%;">
<path d="M9 16.2L4.8 12l-1.4 1.4L9 19 21 7l-1.4-1.4L9 16.2z" />
</svg>`,jn=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg t="1655876154826" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="22" height="22">
<path d="M571.733333 512l268.8-268.8c17.066667-17.066667 17.066667-42.666667 0-59.733333-17.066667-17.066667-42.666667-17.066667-59.733333 0L512 452.266667 243.2 183.466667c-17.066667-17.066667-42.666667-17.066667-59.733333 0-17.066667 17.066667-17.066667 42.666667 0 59.733333L452.266667 512 183.466667 780.8c-17.066667 17.066667-17.066667 42.666667 0 59.733333 8.533333 8.533333 19.2 12.8 29.866666 12.8s21.333333-4.266667 29.866667-12.8L512 571.733333l268.8 268.8c8.533333 8.533333 19.2 12.8 29.866667 12.8s21.333333-4.266667 29.866666-12.8c17.066667-17.066667 17.066667-42.666667 0-59.733333L571.733333 512z" p-id="2131">
</path>
</svg>`,Yn='<svg height="24" viewBox="0 0 24 24" width="24"><path d="M15,17h6v1h-6V17z M11,17H3v1h8v2h1v-2v-1v-2h-1V17z M14,8h1V6V5V3h-1v2H3v1h11V8z            M18,5v1h3V5H18z M6,14h1v-2v-1V9H6v2H3v1 h3V14z M10,12h11v-1H10V12z" data-darkreader-inline-fill="" style="--darkreader-inline-fill:#a8a6a4;"></path></svg>',qn=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg t="1652850026663" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="2749" xmlns:xlink="http://www.w3.org/1999/xlink" width="50" height="50">
<path d="M593.8176 168.5504l356.00384 595.21024c26.15296 43.74528 10.73152 99.7376-34.44736 125.05088-14.39744 8.06912-30.72 12.30848-47.37024 12.30848H155.97568C103.75168 901.12 61.44 860.16 61.44 809.61536c0-16.09728 4.38272-31.92832 12.71808-45.8752L430.16192 168.5504c26.17344-43.7248 84.00896-58.65472 129.20832-33.34144a93.0816 93.0816 0 0 1 34.44736 33.34144zM512 819.2a61.44 61.44 0 1 0 0-122.88 61.44 61.44 0 0 0 0 122.88z m0-512a72.31488 72.31488 0 0 0-71.76192 81.3056l25.72288 205.7216a46.40768 46.40768 0 0 0 92.07808 0l25.72288-205.74208A72.31488 72.31488 0 0 0 512 307.2z" p-id="2750">
</path>
</svg>`,Xn=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg t="1652445277062" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="24" height="24">
<path d="M554.666667 810.666667v85.333333h-85.333334v-85.333333h85.333334zM170.666667 178.005333a42.666667 42.666667 0 0 1 34.986666 18.218667l203.904 291.328a42.666667 42.666667 0 0 1 0 48.896l-203.946666 291.328A42.666667 42.666667 0 0 1 128 803.328V220.672a42.666667 42.666667 0 0 1 42.666667-42.666667z m682.666666 0a42.666667 42.666667 0 0 1 42.368 37.717334l0.298667 4.949333v582.656a42.666667 42.666667 0 0 1-74.24 28.629333l-3.413333-4.181333-203.904-291.328a42.666667 42.666667 0 0 1-3.029334-43.861333l3.029334-5.034667 203.946666-291.328A42.666667 42.666667 0 0 1 853.333333 178.005333zM554.666667 640v85.333333h-85.333334v-85.333333h85.333334zM196.266667 319.104V716.8L335.957333 512 196.309333 319.104zM554.666667 469.333333v85.333334h-85.333334v-85.333334h85.333334z m0-170.666666v85.333333h-85.333334V298.666667h85.333334z m0-170.666667v85.333333h-85.333334V128h85.333334z">
</path>
</svg>
`,Un=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg class="icon" width="22" height="22" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg">
<path d="M768 298.666667h170.666667v85.333333h-256V128h85.333333v170.666667zM341.333333 384H85.333333V298.666667h170.666667V128h85.333333v256z m426.666667 341.333333v170.666667h-85.333333v-256h256v85.333333h-170.666667zM341.333333 640v256H256v-170.666667H85.333333v-85.333333h256z" />
</svg>
`,Gn=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg class="icon" width="22" height="22" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg">
<path d="M625.777778 256h142.222222V398.222222h113.777778V142.222222H625.777778v113.777778zM256 398.222222V256H398.222222v-113.777778H142.222222V398.222222h113.777778zM768 625.777778v142.222222H625.777778v113.777778h256V625.777778h-113.777778zM398.222222 768H256V625.777778h-113.777778v256H398.222222v-113.777778z" />
</svg>
`,Zn=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg class="icon" width="18" height="18" viewBox="0 0 1152 1024" version="1.1" xmlns="http://www.w3.org/2000/svg">
<path d="M1075.2 0H76.8A76.8 76.8 0 0 0 0 76.8v870.4A76.8 76.8 0 0 0 76.8 1024h998.4a76.8 76.8 0 0 0 76.8-76.8V76.8A76.8 76.8 0 0 0 1075.2 0zM1024 128v768H128V128h896zM896 512a64 64 0 0 1 7.488 127.552L896 640h-128v128a64 64 0 0 1-56.512 63.552L704 832a64 64 0 0 1-63.552-56.512L640 768V582.592c0-34.496 25.024-66.112 61.632-70.208L709.632 512H896zM256 512a64 64 0 0 1-7.488-127.552L256 384h128V256a64 64 0 0 1 56.512-63.552L448 192a64 64 0 0 1 63.552 56.512L512 256v185.408c0 34.432-25.024 66.112-61.632 70.144L442.368 512H256z" />
</svg>
`,Kn=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg class="icon" width="18" height="18" viewBox="0 0 1152 1024" version="1.1" xmlns="http://www.w3.org/2000/svg">
<path d="M1075.2 0H76.8A76.8 76.8 0 0 0 0 76.8v870.4A76.8 76.8 0 0 0 76.8 1024h998.4a76.8 76.8 0 0 0 76.8-76.8V76.8A76.8 76.8 0 0 0 1075.2 0zM1024 128v768H128V128h896zM448 192a64 64 0 0 1 7.488 127.552L448 320H320v128a64 64 0 0 1-56.512 63.552L256 512a64 64 0 0 1-63.552-56.512L192 448V262.592c0-34.432 25.024-66.112 61.632-70.144L261.632 192H448zM704 832a64 64 0 0 1-7.488-127.552L704 704h128V576a64 64 0 0 1 56.512-63.552L896 512a64 64 0 0 1 63.552 56.512L960 576v185.408c0 34.496-25.024 66.112-61.632 70.208l-8 0.384H704z" />
</svg>
`,Jn=`<svg xmlns="http://www.w3.org/2000/svg" width="50px" height="50px" viewBox="0 0 100 100" preserveAspectRatio="xMidYMid" class="uil-default">
  <rect x="0" y="0" width="100" height="100" fill="none" class="bk"/>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(0 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-1s" repeatCount="indefinite"/>
  </rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(30 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.9166666666666666s" repeatCount="indefinite"/>
  </rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(60 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.8333333333333334s" repeatCount="indefinite"/>
  </rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(90 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.75s" repeatCount="indefinite"/></rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(120 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.6666666666666666s" repeatCount="indefinite"/>
  </rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(150 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.5833333333333334s" repeatCount="indefinite"/>
  </rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(180 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.5s" repeatCount="indefinite"/></rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(210 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.4166666666666667s" repeatCount="indefinite"/>
  </rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(240 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.3333333333333333s" repeatCount="indefinite"/>
  </rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(270 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.25s" repeatCount="indefinite"/></rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(300 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.16666666666666666s" repeatCount="indefinite"/>
  </rect>
  <rect x="47" y="40" width="6" height="20" rx="5" ry="5" transform="rotate(330 50 50) translate(0 -30)">
    <animate attributeName="opacity" from="1" to="0" dur="1s" begin="-0.08333333333333333s" repeatCount="indefinite"/>
  </rect>
</svg>`,Qn=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg t="1650612139149" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="20" height="20">
<path d="M298.666667 426.666667V341.333333a213.333333 213.333333 0 1 1 426.666666 0v85.333334h42.666667a85.333333 85.333333 0 0 1 85.333333 85.333333v256a85.333333 85.333333 0 0 1-85.333333 85.333333H256a85.333333 85.333333 0 0 1-85.333333-85.333333v-256a85.333333 85.333333 0 0 1 85.333333-85.333333h42.666667z m213.333333-213.333334a128 128 0 0 0-128 128v85.333334h256V341.333333a128 128 0 0 0-128-128z"></path>
</svg>
`,ti=`<svg xmlns="http://www.w3.org/2000/svg" height="22" width="22" viewBox="0 0 22 22">
    <path d="M7 3a2 2 0 0 0-2 2v12a2 2 0 1 0 4 0V5a2 2 0 0 0-2-2zM15 3a2 2 0 0 0-2 2v12a2 2 0 1 0 4 0V5a2 2 0 0 0-2-2z"></path>
</svg>`,ei=`<svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" width="22" height="22">
<path d="M844.8 219.648h-665.6c-6.144 0-10.24 4.608-10.24 10.752v563.2c0 5.632 4.096 10.24 10.24 10.24h256v92.16h-256a102.4 102.4 0 0 1-102.4-102.4v-563.2c0-56.832 45.568-102.4 102.4-102.4h665.6a102.4 102.4 0 0 1 102.4 102.4v204.8h-92.16v-204.8c0-6.144-4.608-10.752-10.24-10.752zM614.4 588.8c-28.672 0-51.2 22.528-51.2 51.2v204.8c0 28.16 22.528 51.2 51.2 51.2h281.6c28.16 0 51.2-23.04 51.2-51.2v-204.8c0-28.672-23.04-51.2-51.2-51.2H614.4z"></path>
</svg>`,ni=`<svg xmlns="http://www.w3.org/2000/svg" height="22" width="22" viewBox="0 0 22 22">
  <path d="M17.982 9.275L8.06 3.27A2.013 2.013 0 0 0 5 4.994v12.011a2.017 2.017 0 0 0 3.06 1.725l9.922-6.005a2.017 2.017 0 0 0 0-3.45z"></path>
</svg>`,ii='<svg height="24" viewBox="0 0 24 24" width="24"><path d="M10,8v8l6-4L10,8L10,8z M6.3,5L5.7,4.2C7.2,3,9,2.2,11,2l0.1,1C9.3,3.2,7.7,3.9,6.3,5z            M5,6.3L4.2,5.7C3,7.2,2.2,9,2,11 l1,.1C3.2,9.3,3.9,7.7,5,6.3z            M5,17.7c-1.1-1.4-1.8-3.1-2-4.8L2,13c0.2,2,1,3.8,2.2,5.4L5,17.7z            M11.1,21c-1.8-0.2-3.4-0.9-4.8-2 l-0.6,.8C7.2,21,9,21.8,11,22L11.1,21z            M22,12c0-5.2-3.9-9.4-9-10l-0.1,1c4.6,.5,8.1,4.3,8.1,9s-3.5,8.5-8.1,9l0.1,1 C18.2,21.5,22,17.2,22,12z" data-darkreader-inline-fill="" style="--darkreader-inline-fill:#a8a6a4;"></path></svg>',oi=`<svg xmlns="http://www.w3.org/2000/svg" height="22" width="22" viewBox="0 0 50 50">
	<path d="M 19.402344 6 C 17.019531 6 14.96875 7.679688 14.5 10.011719 L 14.097656 12 L 9 12 C 6.238281 12 4 14.238281 4 17 L 4 38 C 4 40.761719 6.238281 43 9 43 L 41 43 C 43.761719 43 46 40.761719 46 38 L 46 17 C 46 14.238281 43.761719 12 41 12 L 35.902344 12 L 35.5 10.011719 C 35.03125 7.679688 32.980469 6 30.597656 6 Z M 25 17 C 30.519531 17 35 21.480469 35 27 C 35 32.519531 30.519531 37 25 37 C 19.480469 37 15 32.519531 15 27 C 15 21.480469 19.480469 17 25 17 Z M 25 19 C 20.589844 19 17 22.589844 17 27 C 17 31.410156 20.589844 35 25 35 C 29.410156 35 33 31.410156 33 27 C 33 22.589844 29.410156 19 25 19 Z "/>
</svg>
`,ri=`<svg xmlns="http://www.w3.org/2000/svg" height="22" width="22" viewBox="0 0 22 22">
    <circle cx="11" cy="11" r="2"></circle>
    <path d="M19.164 8.861L17.6 8.6a6.978 6.978 0 0 0-1.186-2.099l.574-1.533a1 1 0 0 0-.436-1.217l-1.997-1.153a1.001 1.001 0 0 0-1.272.23l-1.008 1.225a7.04 7.04 0 0 0-2.55.001L8.716 2.829a1 1 0 0 0-1.272-.23L5.447 3.751a1 1 0 0 0-.436 1.217l.574 1.533A6.997 6.997 0 0 0 4.4 8.6l-1.564.261A.999.999 0 0 0 2 9.847v2.306c0 .489.353.906.836.986l1.613.269a7 7 0 0 0 1.228 2.075l-.558 1.487a1 1 0 0 0 .436 1.217l1.997 1.153c.423.244.961.147 1.272-.23l1.04-1.263a7.089 7.089 0 0 0 2.272 0l1.04 1.263a1 1 0 0 0 1.272.23l1.997-1.153a1 1 0 0 0 .436-1.217l-.557-1.487c.521-.61.94-1.31 1.228-2.075l1.613-.269a.999.999 0 0 0 .835-.986V9.847a.999.999 0 0 0-.836-.986zM11 15a4 4 0 1 1 0-8 4 4 0 0 1 0 8z"></path>
</svg>`,ai=`<svg xmlns="http://www.w3.org/2000/svg" width="80" height="80" viewBox="0 0 24 24">
<path d="M9.5 9.325v5.35q0 .575.525.875t1.025-.05l4.15-2.65q.475-.3.475-.85t-.475-.85L11.05 8.5q-.5-.35-1.025-.05t-.525.875ZM12 22q-2.075 0-3.9-.788t-3.175-2.137q-1.35-1.35-2.137-3.175T2 12q0-2.075.788-3.9t2.137-3.175q1.35-1.35 3.175-2.137T12 2q2.075 0 3.9.788t3.175 2.137q1.35 1.35 2.138 3.175T22 12q0 2.075-.788 3.9t-2.137 3.175q-1.35 1.35-3.175 2.138T12 22Z"/>
</svg>
`,si=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg class="icon" width="26" height="26" viewBox="0 0 1740 1024" version="1.1" xmlns="http://www.w3.org/2000/svg">
    <path d="M511.8976 1024h670.5152c282.4192-0.4096 511.1808-229.4784 511.1808-511.8976 0-282.4192-228.7616-511.488-511.1808-511.8976H511.8976C229.4784 0.6144 0.7168 229.6832 0.7168 512.1024c0 282.4192 228.7616 511.488 511.1808 511.8976zM511.3344 48.64A464.5888 464.5888 0 1 1 48.0256 513.024 463.872 463.872 0 0 1 511.3344 48.4352V48.64z" />
</svg>
`,li=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg class="icon" width="26" height="26" viewBox="0 0 1664 1024" version="1.1" xmlns="http://www.w3.org/2000/svg">
    <path fill="#648FFC" d="M1152 0H512a512 512 0 0 0 0 1024h640a512 512 0 0 0 0-1024z m0 960a448 448 0 1 1 448-448 448 448 0 0 1-448 448z"  />
</svg>`,ci=`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg t="1650612464266" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="20" height="20"><path d="M666.752 194.517333L617.386667 268.629333A128 128 0 0 0 384 341.333333l0.042667 85.333334h384a85.333333 85.333333 0 0 1 85.333333 85.333333v256a85.333333 85.333333 0 0 1-85.333333 85.333333H256a85.333333 85.333333 0 0 1-85.333333-85.333333v-256a85.333333 85.333333 0 0 1 85.333333-85.333333h42.666667V341.333333a213.333333 213.333333 0 0 1 368.085333-146.816z"></path></svg>
`,di=`<svg xmlns="http://www.w3.org/2000/svg" height="22" width="22" viewBox="0 0 22 22">
    <path d="M15 11a3.998 3.998 0 0 0-2-3.465v2.636l1.865 1.865A4.02 4.02 0 0 0 15 11z"></path>
    <path d="M13.583 5.583A5.998 5.998 0 0 1 17 11a6 6 0 0 1-.585 2.587l1.477 1.477a8.001 8.001 0 0 0-3.446-11.286 1 1 0 0 0-.863 1.805zM18.778 18.778l-2.121-2.121-1.414-1.414-1.415-1.415L13 13l-2-2-3.889-3.889-3.889-3.889a.999.999 0 1 0-1.414 1.414L5.172 8H5a2 2 0 0 0-2 2v2a2 2 0 0 0 2 2h1l4.188 3.35a.5.5 0 0 0 .812-.39v-3.131l2.587 2.587-.01.005a1 1 0 0 0 .86 1.806c.215-.102.424-.214.627-.333l2.3 2.3a1.001 1.001 0 0 0 1.414-1.416zM11 5.04a.5.5 0 0 0-.813-.39L8.682 5.854 11 8.172V5.04z"></path>
</svg>`,pi=`<svg xmlns="http://www.w3.org/2000/svg" height="22" width="22" viewBox="0 0 22 22">
    <path d="M10.188 4.65L6 8H5a2 2 0 0 0-2 2v2a2 2 0 0 0 2 2h1l4.188 3.35a.5.5 0 0 0 .812-.39V5.04a.498.498 0 0 0-.812-.39zM14.446 3.778a1 1 0 0 0-.862 1.804 6.002 6.002 0 0 1-.007 10.838 1 1 0 0 0 .86 1.806A8.001 8.001 0 0 0 19 11a8.001 8.001 0 0 0-4.554-7.222z"></path>
    <path d="M15 11a3.998 3.998 0 0 0-2-3.465v6.93A3.998 3.998 0 0 0 15 11z"></path>
</svg>`;class ui{constructor(t){const e={loading:Jn,state:ai,play:ni,pause:ti,check:Wn,volume:pi,volumeClose:di,screenshot:oi,setting:ri,pip:ei,arrowLeft:Hn,arrowRight:Bn,playbackRate:ii,aspectRatio:Nn,config:Yn,lock:Qn,flip:Xn,unlock:ci,fullscreenOff:Un,fullscreenOn:Gn,fullscreenWebOff:Zn,fullscreenWebOn:Kn,switchOn:li,switchOff:si,error:qn,close:jn,airplay:Fn,...t.option.icons};for(const i in e)$(this,i,{get:()=>ve(i,e[i])})}}class hi extends it{constructor(t){super(t),this.name="info",_||this.init()}init(){const{proxy:t,constructor:e,template:{$infoPanel:i,$infoClose:o,$video:r}}=this.art;t(o,"click",()=>{this.show=!1});let a=null;const s=gt("[data-video]",i)||[];this.art.on("destroy",()=>clearTimeout(a));function c(){for(let l=0;l<s.length;l++){const d=s[l],p=r[d.dataset.video],u=typeof p=="number"?p.toFixed(2):p;d.textContent!==u&&(d.textContent=u)}a=setTimeout(c,e.INFO_LOOP_TIME)}c()}}class fi extends it{constructor(t){super(t);const{option:e,template:{$layer:i}}=t;this.name="layer",this.$parent=i;for(let o=0;o<e.layers.length;o++)this.add(e.layers[o])}}class mi extends it{constructor(t){super(t),this.name="loading",k(t.template.$loading,t.icons.loading)}}class gi extends it{constructor(t){super(t),this.name="mask";const{template:e,icons:i,events:o}=t,r=k(e.$state,i.state),a=k(e.$state,i.error);h(a,"display","none"),t.on("destroy",()=>{h(r,"display","none"),h(a,"display",null)}),o.proxy(e.$state,"click",()=>t.play())}}class vi{constructor(t){this.art=t,this.timer=null,t.on("destroy",()=>this.destroy())}destroy(){this.timer&&(clearTimeout(this.timer),this.timer=null)}set show(t){const{constructor:e,template:{$player:i,$noticeInner:o}}=this.art;t?(o.textContent=t instanceof Error?t.message.trim():t,S(i,"art-notice-show"),clearTimeout(this.timer),this.timer=setTimeout(()=>{o.textContent="",F(i,"art-notice-show")},e.NOTICE_TIME)):F(i,"art-notice-show")}get show(){const{template:{$player:t}}=this.art;return t.classList.contains("art-notice-show")}}function yi(n){const{i18n:t,notice:e,proxy:i,template:{$video:o}}=n;let r=!0;window.WebKitPlaybackTargetAvailabilityEvent&&o.webkitShowPlaybackTargetPicker?i(o,"webkitplaybacktargetavailabilitychanged",a=>{switch(a.availability){case"available":r=!0;break;case"not-available":r=!1;break}}):r=!1,$(n,"airplay",{value(){r?(o.webkitShowPlaybackTargetPicker(),n.emit("airplay")):e.show=t.get("AirPlay Not Available")}})}function bi(n){const{i18n:t,notice:e,template:{$video:i,$player:o}}=n;$(n,"aspectRatio",{get(){return o.dataset.aspectRatio||"default"},set(r){if(r||(r="default"),r==="default")h(i,"width",null),h(i,"height",null),h(i,"margin",null),delete o.dataset.aspectRatio;else{const a=r.split(":").map(Number),{clientWidth:s,clientHeight:c}=o,l=s/c,d=a[0]/a[1];l>d?(h(i,"width",`${d*c}px`),h(i,"height","100%"),h(i,"margin","0 auto")):(h(i,"width","100%"),h(i,"height",`${s/d}px`),h(i,"margin","auto 0")),o.dataset.aspectRatio=r}e.show=`${t.get("Aspect Ratio")}: ${r==="default"?t.get("Default"):r}`,n.emit("aspectRatio",r)}})}function xi(n){const{template:{$video:t}}=n;$(n,"attr",{value(e,i){if(i===void 0)return t[e];t[e]=i}})}function wi(n){const{template:{$container:t,$video:e}}=n;$(n,"autoHeight",{value(){const{clientWidth:i}=t,{videoHeight:o,videoWidth:r}=e,a=o*(i/r);h(t,"height",`${a}px`),n.emit("autoHeight",a)}})}function ki(n){const{$container:t,$player:e,$video:i}=n.template;$(n,"autoSize",{value(){const{videoWidth:o,videoHeight:r}=i,{width:a,height:s}=G(t),c=o/r;if(a/s>c){const d=s*c/a*100;h(e,"width",`${d}%`),h(e,"height","100%")}else{const d=a/c/s*100;h(e,"width","100%"),h(e,"height",`${d}%`)}n.emit("autoSize",{width:n.width,height:n.height})}})}function $i(n){const{$player:t}=n.template;$(n,"cssVar",{value(e,i){return i?t.style.setProperty(e,i):getComputedStyle(t).getPropertyValue(e)}})}function Ti(n){const{$video:t}=n.template;$(n,"currentTime",{get:()=>t.currentTime||0,set:e=>{e=Number.parseFloat(e),!Number.isNaN(e)&&(t.currentTime=j(e,0,n.duration))}})}function Ci(n){$(n,"duration",{get:()=>{const{duration:t}=n.template.$video;return t===1/0?0:t||0}})}function Mi(n){const{i18n:t,notice:e,option:i,constructor:o,proxy:r,template:{$player:a,$video:s,$poster:c}}=n;let l=0;for(let d=0;d<ut.events.length;d++)r(s,ut.events[d],p=>{n.emit(`video:${p.type}`,p)});n.on("video:canplay",()=>{l=0,n.loading.show=!1}),n.once("video:canplay",()=>{n.loading.show=!1,n.controls.show=!0,n.mask.show=!0,n.isReady=!0,n.emit("ready")}),n.on("video:ended",()=>{i.loop?(n.seek=0,n.play(),n.controls.show=!1,n.mask.show=!1):(n.controls.show=!0,n.mask.show=!0)}),n.on("video:error",async d=>{l<o.RECONNECT_TIME_MAX?(await at(o.RECONNECT_SLEEP_TIME),l+=1,n.url=i.url,e.show=`${t.get("Reconnect")}: ${l}`,n.emit("error",d,l)):(n.mask.show=!0,n.loading.show=!1,n.controls.show=!0,S(a,"art-error"),await at(o.RECONNECT_SLEEP_TIME),e.show=t.get("Video Load Failed"))}),n.on("video:loadedmetadata",()=>{n.emit("resize"),_&&(n.loading.show=!1,n.controls.show=!0,n.mask.show=!0)}),n.on("video:loadstart",()=>{n.loading.show=!0,n.mask.show=!1,n.controls.show=!0}),n.on("video:pause",()=>{n.controls.show=!0,n.mask.show=!0}),n.on("video:play",()=>{n.mask.show=!1,h(c,"display","none")}),n.on("video:playing",()=>{n.mask.show=!1}),n.on("video:progress",()=>{n.playing&&(n.loading.show=!1)}),n.on("video:seeked",()=>{n.loading.show=!1,n.mask.show=!0}),n.on("video:seeking",()=>{n.loading.show=!0,n.mask.show=!1}),n.on("video:timeupdate",()=>{n.mask.show=!1}),n.on("video:waiting",()=>{n.loading.show=!0,n.mask.show=!1})}function Ei(n){const{template:{$player:t},i18n:e,notice:i}=n;$(n,"flip",{get(){return t.dataset.flip||"normal"},set(o){o||(o="normal"),o==="normal"?delete t.dataset.flip:t.dataset.flip=o,i.show=`${e.get("Video Flip")}: ${e.get(Et(o))}`,n.emit("flip",o)}})}const Qt=[["requestFullscreen","exitFullscreen","fullscreenElement","fullscreenEnabled","fullscreenchange","fullscreenerror"],["webkitRequestFullscreen","webkitExitFullscreen","webkitFullscreenElement","webkitFullscreenEnabled","webkitfullscreenchange","webkitfullscreenerror"],["webkitRequestFullScreen","webkitCancelFullScreen","webkitCurrentFullScreenElement","webkitCancelFullScreen","webkitfullscreenchange","webkitfullscreenerror"],["mozRequestFullScreen","mozCancelFullScreen","mozFullScreenElement","mozFullScreenEnabled","mozfullscreenchange","mozfullscreenerror"],["msRequestFullscreen","msExitFullscreen","msFullscreenElement","msFullscreenEnabled","MSFullscreenChange","MSFullscreenError"]],nt=(()=>{if(typeof document>"u")return!1;const n=Qt[0],t={};for(const e of Qt)if(e[1]in document){for(const[o,r]of e.entries())t[n[o]]=r;return t}return!1})(),te={change:nt.fullscreenchange,error:nt.fullscreenerror},O={request(n=document.documentElement,t){return new Promise((e,i)=>{const o=()=>{O.off("change",o),e()};O.on("change",o);const r=n[nt.requestFullscreen](t);r instanceof Promise&&r.then(o).catch(i)})},exit(){return new Promise((n,t)=>{if(!O.isFullscreen){n();return}const e=()=>{O.off("change",e),n()};O.on("change",e);const i=document[nt.exitFullscreen]();i instanceof Promise&&i.then(e).catch(t)})},toggle(n,t){return O.isFullscreen?O.exit():O.request(n,t)},onchange(n){O.on("change",n)},onerror(n){O.on("error",n)},on(n,t){const e=te[n];e&&document.addEventListener(e,t,!1)},off(n,t){const e=te[n];e&&document.removeEventListener(e,t,!1)},raw:nt};Object.defineProperties(O,{isFullscreen:{get:()=>!!document[nt.fullscreenElement]},element:{enumerable:!0,get:()=>document[nt.fullscreenElement]},isEnabled:{enumerable:!0,get:()=>!!document[nt.fullscreenEnabled]}});function Si(n){const{i18n:t,notice:e,template:{$video:i,$player:o}}=n,r=s=>{O.on("change",()=>{s.emit("fullscreen",O.isFullscreen),O.isFullscreen?(s.state="fullscreen",S(o,"art-fullscreen")):F(o,"art-fullscreen"),s.emit("resize")}),O.on("error",c=>{s.emit("fullscreenError",c)}),$(s,"fullscreen",{get(){return O.isFullscreen},async set(c){c?await O.request(o):await O.exit()}})},a=s=>{s.on("document:webkitfullscreenchange",()=>{s.emit("fullscreen",s.fullscreen),s.emit("resize")}),$(s,"fullscreen",{get(){return document.fullscreenElement===i},set(c){c?(s.state="fullscreen",i.webkitEnterFullscreen()):i.webkitExitFullscreen()}})};n.once("video:loadedmetadata",()=>{O.isEnabled?r(n):i.webkitSupportsFullscreen?a(n):$(n,"fullscreen",{get(){return!1},set(){e.show=t.get("Fullscreen Not Supported")}}),$(n,"fullscreen",Te(n,"fullscreen"))})}function zi(n){const{constructor:t,template:{$container:e,$player:i}}=n;let o="";$(n,"fullscreenWeb",{get(){return Q(i,"art-fullscreen-web")},set(r){r?(o=i.style.cssText,t.FULLSCREEN_WEB_IN_BODY&&k(document.body,i),n.state="fullscreenWeb",h(i,"width","100%"),h(i,"height","100%"),S(i,"art-fullscreen-web"),n.emit("fullscreenWeb",!0)):(t.FULLSCREEN_WEB_IN_BODY&&k(e,i),o&&(i.style.cssText=o,o=""),F(i,"art-fullscreen-web"),n.emit("fullscreenWeb",!1)),n.emit("resize")}})}function Li(n){const{$video:t}=n.template;$(n,"loaded",{get:()=>n.loadedTime/t.duration}),$(n,"loadedTime",{get:()=>t.buffered.length?t.buffered.end(t.buffered.length-1):0})}function Pi(n){const{icons:t,proxy:e,storage:i,template:{$player:o,$video:r}}=n;let a=!1,s=0,c=0;function l(){const{$mini:m}=n.template;m&&(F(o,"art-mini"),h(m,"display","none"),o.prepend(r),n.emit("mini",!1))}function d(m,v){n.playing?(h(m,"display","none"),h(v,"display","flex")):(h(m,"display","flex"),h(v,"display","none"))}function p(){const{$mini:m}=n.template;if(m)return k(m,r),h(m,"display","flex");{const v=D("div");S(v,"art-mini-popup"),k(document.body,v),n.template.$mini=v,k(v,r);const f=k(v,'<div class="art-mini-close"></div>');k(f,t.close),e(f,"click",l);const b=k(v,'<div class="art-mini-state"></div>'),C=k(b,t.play),T=k(b,t.pause);return e(C,"click",()=>n.play()),e(T,"click",()=>n.pause()),d(C,T),n.on("video:playing",()=>d(C,T)),n.on("video:pause",()=>d(C,T)),n.on("video:timeupdate",()=>d(C,T)),e(v,"mousedown",y=>{a=y.button===0,s=y.pageX,c=y.pageY}),n.on("document:mousemove",y=>{if(a){S(v,"art-mini-dragging");const g=y.pageX-s,x=y.pageY-c;h(v,"transform",`translate(${g}px, ${x}px)`)}}),n.on("document:mouseup",()=>{if(a){a=!1,F(v,"art-mini-dragging");const y=G(v);i.set("left",y.left),i.set("top",y.top),h(v,"left",`${y.left}px`),h(v,"top",`${y.top}px`),h(v,"transform",null)}}),v}}function u(){const{$mini:m}=n.template,v=G(m),f=window.innerHeight-v.height-50,b=window.innerWidth-v.width-50;i.set("top",f),i.set("left",b),h(m,"top",`${f}px`),h(m,"left",`${b}px`)}$(n,"mini",{get(){return Q(o,"art-mini")},set(m){if(m){n.state="mini",S(o,"art-mini");const v=p(),f=i.get("top"),b=i.get("left");typeof f=="number"&&typeof b=="number"?(h(v,"top",`${f}px`),h(v,"left",`${b}px`),Bt(v)||u()):u(),n.emit("mini",!0)}else l()}})}function _i(n){const{option:t,storage:e,template:{$video:i,$poster:o}}=n;for(const a in t.moreVideoAttr)n.attr(a,t.moreVideoAttr[a]);t.muted&&(n.muted=t.muted),t.volume&&(i.volume=j(t.volume,0,1));const r=e.get("volume");typeof r=="number"&&(i.volume=j(r,0,1)),t.poster&&h(o,"backgroundImage",`url(${t.poster})`),t.autoplay&&(i.autoplay=t.autoplay),t.playsInline&&(i.playsInline=!0,i["webkit-playsinline"]=!0),t.theme&&(t.cssVar["--art-theme"]=t.theme);for(const a in t.cssVar)n.cssVar(a,t.cssVar[a]);n.url=t.url}function Ii(n){const{template:{$video:t},i18n:e,notice:i}=n;$(n,"pause",{value(){const o=t.pause();return i.show=e.get("Pause"),n.emit("pause"),o}})}function Ai(n){const{template:{$video:t},proxy:e,notice:i}=n;t.disablePictureInPicture=!1,$(n,"pip",{get(){return document.pictureInPictureElement},set(o){o?(n.state="pip",t.requestPictureInPicture().catch(r=>{throw i.show=r,r})):document.exitPictureInPicture().catch(r=>{throw i.show=r,r})}}),e(t,"enterpictureinpicture",()=>{n.emit("pip",!0)}),e(t,"leavepictureinpicture",()=>{n.emit("pip",!1)})}function Ri(n){const{$video:t}=n.template;t.webkitSetPresentationMode("inline"),$(n,"pip",{get(){return t.webkitPresentationMode==="picture-in-picture"},set(e){e?(n.state="pip",t.webkitSetPresentationMode("picture-in-picture"),n.emit("pip",!0)):(t.webkitSetPresentationMode("inline"),n.emit("pip",!1))}})}function Vi(n){const{i18n:t,notice:e,template:{$video:i}}=n;document.pictureInPictureEnabled?Ai(n):i.webkitSupportsPresentationMode?Ri(n):$(n,"pip",{get(){return!1},set(){e.show=t.get("PIP Not Supported")}})}function Di(n){const{template:{$video:t},i18n:e,notice:i}=n;$(n,"playbackRate",{get(){return t.playbackRate},set(o){if(o){if(o===t.playbackRate)return;t.playbackRate=o,i.show=`${e.get("Rate")}: ${o===1?e.get("Normal"):`${o}x`}`}else n.playbackRate=1}})}function Oi(n){$(n,"played",{get:()=>n.currentTime/n.duration})}function Fi(n){const{$video:t}=n.template;$(n,"playing",{get:()=>typeof t.playing=="boolean"?t.playing:t.currentTime>0&&!t.paused&&!t.ended&&t.readyState>2})}function Hi(n){const{i18n:t,notice:e,option:i,constructor:{instances:o},template:{$video:r}}=n;$(n,"play",{async value(){const a=await r.play();if(e.show=t.get("Play"),n.emit("play"),i.mutex)for(let s=0;s<o.length;s++){const c=o[s];c!==n&&c.pause()}return a}})}function Bi(n){const{template:{$poster:t}}=n;$(n,"poster",{get:()=>{try{return t.style.backgroundImage.match(/"(.*)"/)[1]}catch{return""}},set(e){h(t,"backgroundImage",`url(${e})`)}})}function Ni(n){$(n,"quality",{set(t){const{controls:e,notice:i,i18n:o}=n,r=t.find(a=>a.default)||t[0];e.update({name:"quality",position:"right",index:10,style:{marginRight:"10px"},html:r?.html||"",selector:t,async onSelect(a){return await n.switchQuality(a.url),i.show=`${o.get("Switch Video")}: ${a.html}`,a.html}})}})}function Wi(n){$(n,"rect",{get:()=>G(n.template.$player)});const t=["bottom","height","left","right","top","width"];for(let e=0;e<t.length;e++){const i=t[e];$(n,i,{get:()=>n.rect[i]})}$(n,"x",{get:()=>n.left+window.pageXOffset}),$(n,"y",{get:()=>n.top+window.pageYOffset})}function ji(n){const{notice:t,template:{$video:e}}=n,i=D("canvas");$(n,"getDataURL",{value:()=>new Promise((o,r)=>{try{i.width=e.videoWidth,i.height=e.videoHeight,i.getContext("2d").drawImage(e,0,0),o(i.toDataURL("image/png"))}catch(a){t.show=a,r(a)}})}),$(n,"getBlobUrl",{value:()=>new Promise((o,r)=>{try{i.width=e.videoWidth,i.height=e.videoHeight,i.getContext("2d").drawImage(e,0,0),i.toBlob(a=>{o(URL.createObjectURL(a))})}catch(a){t.show=a,r(a)}})}),$(n,"screenshot",{value:async o=>{const r=await n.getDataURL(),a=o||`artplayer_${Z(e.currentTime)}`;return ke(r,`${a}.png`),n.emit("screenshot",r),r}})}function Yi(n){const{notice:t}=n;$(n,"seek",{set(e){n.currentTime=e,n.duration&&(t.show=`${Z(n.currentTime)} / ${Z(n.duration)}`),n.emit("seek",n.currentTime,e)}}),$(n,"forward",{set(e){n.seek=n.currentTime+e}}),$(n,"backward",{set(e){n.seek=n.currentTime-e}})}function qi(n){const t=["mini","pip","fullscreen","fullscreenWeb"];$(n,"state",{get:()=>t.find(e=>n[e])||"standard",set(e){for(let i=0;i<t.length;i++){const o=t[i];o!==e&&n[o]&&(n[o]=!1)}}})}function Xi(n){const{notice:t,i18n:e,template:i}=n;$(n,"subtitleOffset",{get(){return i.$track?.offset||0},set(o){const{cues:r}=n.subtitle;if(!i.$track||r.length===0)return;const a=j(o,-10,10);i.$track.offset=a;for(let s=0;s<r.length;s++){const c=r[s];c.originalStartTime=c.originalStartTime??c.startTime,c.originalEndTime=c.originalEndTime??c.endTime,c.startTime=j(c.originalStartTime+a,0,n.duration),c.endTime=j(c.originalEndTime+a,0,n.duration)}n.subtitle.update(),t.show=`${e.get("Subtitle Offset")}: ${o}s`,n.emit("subtitleOffset",o)}})}function Ui(n){function t(e,i){return new Promise((o,r)=>{if(e===n.url){o();return}const{playing:a,aspectRatio:s,playbackRate:c}=n;n.pause(),n.url=e,n.notice.show="";const l={};l.error=d=>{n.off("video:canplay",l.canplay),n.off("video:loadedmetadata",l.metadata),r(d)},l.metadata=()=>{n.currentTime=i},l.canplay=async()=>{n.off("video:error",l.error),n.playbackRate=c,n.aspectRatio=s,a&&await n.play(),n.notice.show="",o()},n.once("video:error",l.error),n.once("video:loadedmetadata",l.metadata),n.once("video:canplay",l.canplay)})}$(n,"switchQuality",{value:e=>t(e,n.currentTime)}),$(n,"switchUrl",{value:e=>t(e,0)}),$(n,"switch",{set:n.switchUrl})}function Gi(n){$(n,"theme",{get(){return n.cssVar("--art-theme")},set(t){n.cssVar("--art-theme",t)}})}function Zi(n){const{option:t,template:{$progress:e,$video:i}}=n;let o=null,r=!1,a=null;function s(){clearTimeout(o),o=null,r=!1,a=null}function c(l){const d=n.controls?.thumbnails;if(!d)return;const{number:p,column:u,width:m,height:v,scale:f}=t.thumbnails,b=m*f||a.naturalWidth/u,C=v*f||b/(i.videoWidth/i.videoHeight),T=e.clientWidth/p,y=Math.floor(l/T),g=Math.ceil(y/u)-1,x=y%u||u-1;h(d,"backgroundImage",`url(${a.src})`),h(d,"height",`${C}px`),h(d,"width",`${b}px`),h(d,"backgroundPosition",`-${x*b}px -${g*C}px`),l<=b/2?h(d,"left",0):l>e.clientWidth-b/2?h(d,"left",`${e.clientWidth-b}px`):h(d,"left",`${l-b/2}px`)}n.on("setBar",async(l,d,p)=>{const u=n.controls?.thumbnails,{url:m,scale:v}=t.thumbnails;if(!u||!m)return;if(l==="hover"||l==="played"&&p&&_){if(!a&&!r&&(r=!0,a=await xe(m,v),r=!1),!a)return;const b=e.clientWidth*d;b>0&&b<e.clientWidth&&c(b)}}),$(n,"thumbnails",{get(){return n.option.thumbnails},set(l){l.url&&!n.option.isLive&&(n.option.thumbnails=l,s())}})}function Ki(n){$(n,"toggle",{value(){return n.playing?n.pause():n.play()}})}function Ji(n){$(n,"type",{get(){return n.option.type},set(t){n.option.type=t}})}function Qi(n){const{option:t,template:{$video:e}}=n;$(n,"url",{get(){return e.src},async set(i){if(i){const o=n.url,r=t.type||ft(i),a=t.customType[r];r&&a?(await at(),n.loading.show=!0,a.call(n,e,i,n)):(URL.revokeObjectURL(o),e.src=i),o!==n.url&&(n.option.url=i,n.isReady&&o&&n.once("video:canplay",()=>{n.emit("restart",i)}))}else await at(),n.loading.show=!0}})}function to(n){const{template:{$video:t},i18n:e,notice:i,storage:o}=n;$(n,"volume",{get:()=>t.volume||0,set:r=>{t.volume=j(r,0,1),i.show=`${e.get("Volume")}: ${Number.parseInt(t.volume*100,10)}`,t.volume!==0&&o.set("volume",t.volume)}}),$(n,"muted",{get:()=>t.muted,set:r=>{t.muted=r,n.emit("muted",r)}})}class eo{constructor(t){Qi(t),xi(t),Hi(t),Ii(t),Ki(t),Yi(t),to(t),Ti(t),Ci(t),Ui(t),Di(t),bi(t),ji(t),Si(t),zi(t),Vi(t),Li(t),Oi(t),Fi(t),ki(t),Wi(t),Ei(t),Pi(t),Bi(t),wi(t),$i(t),Gi(t),Ji(t),qi(t),Xi(t),yi(t),Ni(t),Zi(t),Mi(t),_i(t)}}function no(n){const{notice:t,constructor:e,template:{$player:i,$video:o}}=n,r="art-auto-orientation",a="art-auto-orientation-fullscreen";let s=!1;function c(){const p=document.documentElement.clientWidth,u=document.documentElement.clientHeight;h(i,"width",`${u}px`),h(i,"height",`${p}px`),h(i,"transform-origin","0 0"),h(i,"transform",`rotate(90deg) translate(0, -${p}px)`),S(i,r),n.isRotate=!0,n.emit("resize")}function l(){h(i,"width",""),h(i,"height",""),h(i,"transform-origin",""),h(i,"transform",""),F(i,r),n.isRotate=!1,n.emit("resize")}function d(){const{videoWidth:p,videoHeight:u}=o,m=document.documentElement.clientWidth,v=document.documentElement.clientHeight;return p>u&&m<v||p<u&&m>v}return n.on("fullscreenWeb",p=>{if(p){if(d()){const u=Number(e.AUTO_ORIENTATION_TIME??0);setTimeout(()=>{n.fullscreenWeb&&!Q(i,r)&&c()},u)}}else Q(i,r)&&l()}),n.on("fullscreen",async p=>{const u=!!screen?.orientation?.lock;if(p){if(u&&d())try{const v=screen.orientation.type.startsWith("portrait")?"landscape":"portrait";await screen.orientation.lock(v),s=!0,S(i,a)}catch(m){s=!1,t.show=m}}else if(Q(i,a)&&F(i,a),u&&s){try{screen.orientation.unlock()}catch{}s=!1}}),{name:"autoOrientation",get state(){return Q(i,r)}}}function io(n){const{i18n:t,icons:e,storage:i,constructor:o,proxy:r,template:{$poster:a}}=n,s=n.layers.add({name:"auto-playback",html:`
            <div class="art-auto-playback-close"></div>
            <div class="art-auto-playback-last"></div>
            <div class="art-auto-playback-jump"></div>
        `}),c=W(".art-auto-playback-last",s),l=W(".art-auto-playback-jump",s),d=W(".art-auto-playback-close",s);k(d,e.close);let p=null;n.on("video:timeupdate",()=>{if(n.playing){const m=i.get("times")||{},v=Object.keys(m);v.length>o.AUTO_PLAYBACK_MAX&&delete m[v[0]],m[n.option.id||n.option.url]=n.currentTime,i.set("times",m)}});function u(){const v=(i.get("times")||{})[n.option.id||n.option.url];clearTimeout(p),h(s,"display","none"),v&&v>=o.AUTO_PLAYBACK_MIN&&(h(s,"display","flex"),c.textContent=`${t.get("Last Seen")} ${Z(v)}`,l.textContent=t.get("Jump Play"),r(d,"click",()=>{h(s,"display","none")}),r(l,"click",()=>{n.seek=v,n.play(),h(a,"display","none"),h(s,"display","none")}),n.once("video:timeupdate",()=>{p=setTimeout(()=>{h(s,"display","none")},o.AUTO_PLAYBACK_TIMEOUT)}))}return n.on("ready",u),n.on("restart",u),{name:"auto-playback",get times(){return i.get("times")||{}},clear(){return i.del("times")},delete(m){const v=i.get("times")||{};return delete v[m],i.set("times",v),v}}}function oo(n){const{constructor:t,proxy:e,template:{$player:i,$video:o}}=n;let r=null,a=!1,s=1;const c=d=>{d.touches.length===1&&n.playing&&!n.isLock&&(r=setTimeout(()=>{a=!0,s=n.playbackRate,n.playbackRate=t.FAST_FORWARD_VALUE,S(i,"art-fast-forward")},t.FAST_FORWARD_TIME))},l=()=>{clearTimeout(r),a&&(a=!1,n.playbackRate=s,F(i,"art-fast-forward"))};return e(o,"touchstart",c),n.on("document:touchmove",l),n.on("document:touchend",l),{name:"fastForward",get state(){return Q(i,"art-fast-forward")}}}function ro(n){const{layers:t,icons:e,template:{$player:i}}=n;function o(){return Q(i,"art-lock")}function r(){S(i,"art-lock"),n.isLock=!0,n.emit("lock",!0)}function a(){F(i,"art-lock"),n.isLock=!1,n.emit("lock",!1)}return t.add({name:"lock",mounted(s){const c=k(s,e.lock),l=k(s,e.unlock);h(c,"display","none"),n.on("lock",d=>{d?(h(c,"display","inline-flex"),h(l,"display","none")):(h(c,"display","none"),h(l,"display","inline-flex"))})},click(){o()?a():r()}}),{name:"lock",get state(){return o()},set state(s){s?r():a()}}}function ao(n){return n.on("control",t=>{t?F(n.template.$player,"art-mini-progress-bar"):S(n.template.$player,"art-mini-progress-bar")}),{name:"mini-progress-bar"}}class so{constructor(t){this.art=t,this.id=0;const{option:e}=t;e.miniProgressBar&&!e.isLive&&this.add(ao),e.lock&&_&&this.add(ro),e.autoPlayback&&!e.isLive&&this.add(io),e.autoOrientation&&_&&this.add(no),e.fastForward&&_&&!e.isLive&&this.add(oo);for(let i=0;i<e.plugins.length;i++)this.add(e.plugins[i])}add(t){this.id+=1;const e=t.call(this.art,this.art);return e instanceof Promise?e.then(i=>this.next(t,i)):this.next(t,e)}next(t,e){const i=e&&e.name||t.name||`plugin${this.id}`;return q(!ht(this,i),`Cannot add a plugin that already has the same name: ${i}`),$(this,i,{value:e}),this}}function lo(n){const{i18n:t,icons:e,constructor:{SETTING_ITEM_WIDTH:i,ASPECT_RATIO:o}}=n;function r(s){return s==="default"?t.get("Default"):s}function a(){const s=n.setting.find(`aspect-ratio-${n.aspectRatio}`);n.setting.check(s)}return{width:i,name:"aspect-ratio",html:t.get("Aspect Ratio"),icon:e.aspectRatio,tooltip:r(n.aspectRatio),selector:o.map(s=>({value:s,name:`aspect-ratio-${s}`,default:s===n.aspectRatio,html:r(s)})),onSelect(s){return n.aspectRatio=s.value,s.html},mounted:()=>{a(),n.on("aspectRatio",()=>a())}}}function co(n){const{i18n:t,icons:e,constructor:{SETTING_ITEM_WIDTH:i,FLIP:o}}=n;function r(s){return t.get(Et(s))}function a(){const s=n.setting.find(`flip-${n.flip}`);n.setting.check(s)}return{width:i,name:"flip",html:t.get("Video Flip"),tooltip:r(n.flip),icon:e.flip,selector:o.map(s=>({value:s,name:`flip-${s}`,default:s===n.flip,html:r(s)})),onSelect(s){return n.flip=s.value,s.html},mounted:()=>{a(),n.on("flip",()=>a())}}}function po(n){const{i18n:t,icons:e,constructor:{SETTING_ITEM_WIDTH:i,PLAYBACK_RATE:o}}=n;function r(s){return s===1?t.get("Normal"):s.toFixed(1)}function a(){const s=n.setting.find(`playback-rate-${n.playbackRate}`);n.setting.check(s)}return{width:i,name:"playback-rate",html:t.get("Play Speed"),tooltip:r(n.playbackRate),icon:e.playbackRate,selector:o.map(s=>({value:s,name:`playback-rate-${s}`,default:s===n.playbackRate,html:r(s)})),onSelect(s){return n.playbackRate=s.value,s.html},mounted:()=>{a(),n.on("video:ratechange",()=>a())}}}function uo(n){const{i18n:t,icons:e,constructor:i}=n;return{width:i.SETTING_ITEM_WIDTH,name:"subtitle-offset",html:t.get("Subtitle Offset"),icon:e.subtitle,tooltip:"0s",range:[0,-10,10,.1],onChange(o){return n.subtitleOffset=o.range[0],`${o.range[0]}s`},mounted:(o,r)=>{n.on("subtitleOffset",a=>{r.$range.value=a,r.tooltip=`${a}s`})}}}let ho=class extends it{constructor(t){super(t);const{option:e,controls:i,template:{$setting:o}}=t;this.name="setting",this.$parent=o,this.id=0,this.active=null,this.cache=new Map,this.option=[...this.builtin,...e.settings],e.setting&&(this.format(),this.render(),t.on("blur",()=>{this.show&&(this.show=!1,this.render())}),t.on("focus",r=>{const a=rt(r,i.setting),s=rt(r,this.$parent);this.show&&!a&&!s&&(this.show=!1,this.render())}),t.on("resize",()=>this.resize()))}get builtin(){const t=[],{option:e}=this.art;return e.playbackRate&&t.push(po(this.art)),e.aspectRatio&&t.push(lo(this.art)),e.flip&&t.push(co(this.art)),e.subtitleOffset&&t.push(uo(this.art)),t}traverse(t,e=this.option){for(let i=0;i<e.length;i++){const o=e[i];t(o),o.selector?.length&&this.traverse(t,o.selector)}}check(t){t&&(t.$parent.tooltip=t.html,this.traverse(e=>{e.default=e===t,e.default&&e.$item&&U(e.$item,"art-current")},t.$option),this.render(t.$parents))}format(t=this.option,e,i,o=[]){for(let r=0;r<t.length;r++){const a=t[r];if(a?.name?(q(!o.includes(a.name),`The [${a.name}] already exists in [setting]`),o.push(a.name)):a.name=`setting-${this.id++}`,!a.$formatted){$(a,"$parent",{get:()=>e}),$(a,"$parents",{get:()=>i}),$(a,"$option",{get:()=>t});const s=[];$(a,"$events",{get:()=>s}),$(a,"$formatted",{get:()=>!0})}this.format(a.selector||[],a,t,o)}this.option=t}find(t=""){let e=null;return this.traverse(i=>{i.name===t&&(e=i)}),e}resize(){const{controls:t,constructor:{SETTING_WIDTH:e,SETTING_ITEM_HEIGHT:i},template:{$player:o,$setting:r}}=this.art;if(t.setting&&this.show){const a=this.active[0]?.$parent?.width||e,{left:s,width:c}=G(t.setting),{left:l,width:d}=G(o),p=s-l+c/2-a/2,u=this.active===this.option?this.active.length*i:(this.active.length+1)*i;if(h(r,"height",`${u}px`),h(r,"width",`${a}px`),this.art.isRotate||_)return;p+a>d?(h(r,"left",null),h(r,"right",null)):(h(r,"left",`${p}px`),h(r,"right","auto"))}}inactivate(t){for(let e=0;e<t.$events.length;e++)this.art.events.remove(t.$events[e]);t.$events.length=0}remove(t){const e=this.find(t);q(e,`Can't find [${t}] in the [setting]`);const i=e.$option.indexOf(e);e.$option.splice(i,1),this.inactivate(e),e.$item&&Ct(e.$item),this.render()}update(t){const e=this.find(t.name);return e?(this.inactivate(e),Object.assign(e,t),this.format(),this.createItem(e,!0),this.render(),e):this.add(t)}add(t,e=this.option){return e.push(t),this.format(),this.createItem(t),this.render(),t}createHeader(t){if(!this.cache.has(t.$option))return;const e=this.cache.get(t.$option),{proxy:i,icons:{arrowLeft:o},constructor:{SETTING_ITEM_HEIGHT:r}}=this.art,a=D("div");h(a,"height",`${r}px`),S(a,"art-setting-item"),S(a,"art-setting-item-back");const s=k(a,'<div class="art-setting-item-left"></div>'),c=D("div");S(c,"art-setting-item-left-icon"),k(c,o),k(s,c),k(s,t.$parent.html);const l=i(a,"click",()=>this.render(t.$parents));t.$parent.$events.push(l),k(e,a)}createItem(t,e=!1){if(!this.cache.has(t.$option))return;const i=this.cache.get(t.$option),o=t.$item;let r="selector";ht(t,"switch")&&(r="switch"),ht(t,"range")&&(r="range"),ht(t,"onClick")&&(r="button");const{icons:a,proxy:s,constructor:c}=this.art,l=D("div");S(l,"art-setting-item"),h(l,"height",`${c.SETTING_ITEM_HEIGHT}px`),l.dataset.name=t.name||"",l.dataset.value=t.value||"";const d=k(l,'<div class="art-setting-item-left"></div>'),p=k(l,'<div class="art-setting-item-right"></div>'),u=D("div");switch(S(u,"art-setting-item-left-icon"),r){case"button":case"switch":case"range":k(u,t.icon||a.config);break;case"selector":t.selector?.length?k(u,t.icon||a.config):k(u,a.check);break}k(d,u),$(t,"$icon",{configurable:!0,get:()=>u}),$(t,"icon",{configurable:!0,get(){return u.innerHTML},set(f){u.innerHTML="",k(u,f)}});const m=D("div");S(m,"art-setting-item-left-text"),k(m,t.html||""),k(d,m),$(t,"$html",{configurable:!0,get:()=>m}),$(t,"html",{configurable:!0,get(){return m.innerHTML},set(f){m.innerHTML="",k(m,f)}});const v=D("div");switch(S(v,"art-setting-item-right-tooltip"),k(v,t.tooltip||""),k(p,v),$(t,"$tooltip",{configurable:!0,get:()=>v}),$(t,"tooltip",{configurable:!0,get(){return v.innerHTML},set(f){v.innerHTML="",k(v,f)}}),r){case"switch":{const f=D("div");S(f,"art-setting-item-right-icon");const b=k(f,a.switchOn),C=k(f,a.switchOff);h(t.switch?C:b,"display","none"),k(p,f),$(t,"$switch",{configurable:!0,get:()=>f});let T=t.switch;$(t,"switch",{configurable:!0,get:()=>T,set(y){T=y,y?(h(C,"display","none"),h(b,"display",null)):(h(C,"display",null),h(b,"display","none"))}});break}case"range":{const f=D("div");S(f,"art-setting-item-right-icon");const b=k(f,'<input type="range">');b.value=t.range[0],b.min=t.range[1],b.max=t.range[2],b.step=t.range[3],S(b,"art-setting-range"),k(p,f),$(t,"$range",{configurable:!0,get:()=>b});let C=[...t.range];$(t,"range",{configurable:!0,get:()=>C,set(T){C=[...T],b.value=T[0],b.min=T[1],b.max=T[2],b.step=T[3]}})}break;case"selector":if(t.selector?.length){const f=D("div");S(f,"art-setting-item-right-icon"),k(f,a.arrowRight),k(p,f)}break}switch(r){case"switch":{if(t.onSwitch){const f=s(l,"click",async b=>{t.switch=await t.onSwitch.call(this.art,t,l,b)});t.$events.push(f)}break}case"range":{if(t.$range){if(t.onRange){const f=s(t.$range,"change",async b=>{t.range[0]=t.$range.valueAsNumber,t.tooltip=await t.onRange.call(this.art,t,l,b)});t.$events.push(f)}if(t.onChange){const f=s(t.$range,"input",async b=>{t.range[0]=t.$range.valueAsNumber,t.tooltip=await t.onChange.call(this.art,t,l,b)});t.$events.push(f)}}break}case"selector":{const f=s(l,"click",async b=>{t.selector?.length?this.render(t.selector):(this.check(t),t.$parent.onSelect&&(t.$parent.tooltip=await t.$parent.onSelect.call(this.art,t,l,b)))});t.$events.push(f),t.default&&S(l,"art-current")}break;case"button":if(t.onClick){const f=s(l,"click",async b=>{t.tooltip=await t.onClick.call(this.art,t,l,b)});t.$events.push(f)}break}$(t,"$item",{configurable:!0,get:()=>l}),e?Nt(l,o):k(i,l),t.mounted&&setTimeout(()=>t.mounted.call(this.art,t.$item,t),0)}render(t=this.option){if(this.active=t,this.cache.has(t)){const e=this.cache.get(t);U(e,"art-current")}else{const e=D("div");this.cache.set(t,e),S(e,"art-setting-panel"),k(this.$parent,e),U(e,"art-current"),t[0]?.$parent&&this.createHeader(t[0]);for(let i=0;i<t.length;i++)this.createItem(t[i])}this.resize()}};class fo{constructor(){this.name="artplayer_settings",this.settings={}}get(t){try{const e=JSON.parse(window.localStorage.getItem(this.name))||{};return t?e[t]:e}catch{return t?this.settings[t]:this.settings}}set(t,e){try{const i=Object.assign({},this.get(),{[t]:e});window.localStorage.setItem(this.name,JSON.stringify(i))}catch{this.settings[t]=e}}del(t){try{const e=this.get();delete e[t],window.localStorage.setItem(this.name,JSON.stringify(e))}catch{delete this.settings[t]}}clear(){try{window.localStorage.removeItem(this.name)}catch{this.settings={}}}}const _e=`.art-video-player {
  --art-theme: #f00;
  --art-font-color: #fff;
  --art-background-color: #000;
  --art-text-shadow-color: rgba(0, 0, 0, 0.5);
  --art-transition-duration: 0.2s;
  --art-padding: 10px;
  --art-border-radius: 3px;
  --art-progress-height: 6px;
  --art-progress-color: rgba(255, 255, 255, 0.25);
  --art-progress-top-gap: 10px;
  --art-hover-color: rgba(255, 255, 255, 0.25);
  --art-loaded-color: rgba(255, 255, 255, 0.25);
  --art-state-size: 80px;
  --art-state-opacity: 0.8;
  --art-bottom-height: 100px;
  --art-bottom-offset: 20px;
  --art-bottom-gap: 5px;
  --art-highlight-width: 8px;
  --art-highlight-color: rgba(255, 255, 255, 0.5);
  --art-control-height: 46px;
  --art-control-opacity: 0.75;
  --art-control-icon-size: 36px;
  --art-control-icon-scale: 1.1;
  --art-volume-height: 120px;
  --art-volume-handle-size: 14px;
  --art-lock-size: 36px;
  --art-indicator-scale: 0;
  --art-indicator-size: 16px;
  --art-fullscreen-web-index: 9999;
  --art-settings-icon-size: 24px;
  --art-settings-max-height: 300px;
  --art-selector-max-height: 300px;
  --art-contextmenus-min-width: 250px;
  --art-subtitle-font-size: 20px;
  --art-subtitle-gap: 5px;
  --art-subtitle-bottom: 15px;
  --art-subtitle-border: #000;
  --art-widget-background: rgba(0, 0, 0, 0.85);
  --art-tip-background: rgba(0, 0, 0, 0.7);
  --art-scrollbar-size: 4px;
  --art-scrollbar-background: rgba(255, 255, 255, 0.25);
  --art-scrollbar-background-hover: rgba(255, 255, 255, 0.5);
  --art-mini-progress-height: 2px;
}
.art-bg-cover {
  background-position: center center;
  background-repeat: no-repeat;
  background-size: cover;
}
.art-bottom-gradient {
  background-image: linear-gradient(to top, #000, rgba(0, 0, 0, 0.4), transparent);
  background-repeat: repeat-x;
  background-position: center bottom;
}
.art-backdrop-filter {
  -webkit-backdrop-filter: saturate(180%) blur(20px);
  backdrop-filter: saturate(180%) blur(20px);
  background-color: rgba(0, 0, 0, 0.75) !important;
}
.art-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.art-video-player {
  position: relative;
  margin: 0 auto;
  width: 100%;
  height: 100%;
  outline: 0;
  zoom: 1;
  padding: 0;
  text-align: left;
  direction: ltr;
  font-size: 14px;
  line-height: 1.3;
  user-select: none;
  box-sizing: border-box;
  color: var(--art-font-color);
  background-color: var(--art-background-color);
  text-shadow: 0 0 2px var(--art-text-shadow-color);
  font-family: PingFang SC, Helvetica Neue, Microsoft YaHei, Roboto, Arial, sans-serif;
  -webkit-tap-highlight-color: rgba(0, 0, 0, 0);
  -ms-touch-action: manipulation;
  touch-action: manipulation;
  -ms-high-contrast-adjust: none;
}
.art-video-player *,
.art-video-player *::before,
.art-video-player *::after {
  box-sizing: border-box;
}
.art-video-player ::-webkit-scrollbar {
  width: var(--art-scrollbar-size);
  height: var(--art-scrollbar-size);
}
.art-video-player ::-webkit-scrollbar-thumb {
  background-color: var(--art-scrollbar-background);
}
.art-video-player ::-webkit-scrollbar-thumb:hover {
  background-color: var(--art-scrollbar-background-hover);
}
.art-video-player img {
  max-width: 100%;
  vertical-align: top;
}
.art-video-player svg {
  fill: var(--art-font-color);
}
.art-video-player a {
  color: var(--art-font-color);
  text-decoration: none;
}
.art-icon {
  line-height: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}
.art-video-player.art-backdrop .art-contextmenus,
.art-video-player.art-backdrop .art-info,
.art-video-player.art-backdrop .art-settings,
.art-video-player.art-backdrop .art-layer-auto-playback,
.art-video-player.art-backdrop .art-selector-list,
.art-video-player.art-backdrop .art-volume-inner {
  -webkit-backdrop-filter: saturate(180%) blur(20px);
  backdrop-filter: saturate(180%) blur(20px);
  background-color: rgba(0, 0, 0, 0.75) !important;
}
.art-video {
  position: absolute;
  inset: 0;
  z-index: 10;
  width: 100%;
  height: 100%;
}
.art-poster {
  position: absolute;
  inset: 0;
  z-index: 11;
  width: 100%;
  height: 100%;
  background-position: center center;
  background-repeat: no-repeat;
  background-size: cover;
  pointer-events: none;
}
.art-video-player .art-subtitle {
  display: none;
  justify-content: center;
  align-items: center;
  flex-direction: column;
  position: absolute;
  z-index: 20;
  width: 100%;
  padding: 0 5%;
  text-align: center;
  pointer-events: none;
  gap: var(--art-subtitle-gap);
  bottom: var(--art-subtitle-bottom);
  font-size: var(--art-subtitle-font-size);
  transition: bottom var(--art-transition-duration) ease;
  text-shadow: var(--art-subtitle-border) 1px 0 1px, var(--art-subtitle-border) 0 1px 1px, var(--art-subtitle-border) -1px 0 1px, var(--art-subtitle-border) 0 -1px 1px, var(--art-subtitle-border) 1px 1px 1px, var(--art-subtitle-border) -1px -1px 1px, var(--art-subtitle-border) 1px -1px 1px, var(--art-subtitle-border) -1px 1px 1px;
}
.art-video-player.art-subtitle-show .art-subtitle {
  display: flex;
}
.art-video-player.art-control-show .art-subtitle {
  bottom: calc(var(--art-control-height) + var(--art-subtitle-bottom));
}
.art-danmuku {
  position: absolute;
  inset: 0;
  z-index: 30;
  width: 100%;
  height: 100%;
  pointer-events: none;
  overflow: hidden;
}
.art-video-player .art-layers {
  position: absolute;
  inset: 0;
  z-index: 40;
  width: 100%;
  height: 100%;
  display: none;
  pointer-events: none;
}
.art-video-player .art-layers .art-layer {
  pointer-events: auto;
}
.art-video-player.art-layer-show .art-layers {
  display: flex;
}
.art-video-player .art-mask {
  display: flex;
  justify-content: center;
  align-items: center;
  position: absolute;
  inset: 0;
  z-index: 50;
  width: 100%;
  height: 100%;
  pointer-events: none;
}
.art-video-player .art-mask .art-state {
  display: flex;
  justify-content: center;
  align-items: center;
  opacity: 0;
  transform: scale(2);
  width: var(--art-state-size);
  height: var(--art-state-size);
  transition: all var(--art-transition-duration) ease;
}
.art-video-player.art-mask-show .art-state {
  pointer-events: auto;
  opacity: var(--art-state-opacity);
  transform: scale(1);
}
.art-video-player.art-loading-show .art-state {
  display: none;
}
.art-video-player .art-loading {
  display: none;
  justify-content: center;
  align-items: center;
  position: absolute;
  inset: 0;
  z-index: 70;
  width: 100%;
  height: 100%;
  pointer-events: none;
}
.art-video-player.art-loading-show .art-loading {
  display: flex;
}
.art-video-player.art-loading-show .art-mask {
  display: none;
}
.art-video-player .art-bottom {
  position: absolute;
  inset: 0;
  z-index: 60;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  opacity: 0;
  overflow: hidden;
  pointer-events: none;
  padding: 0 var(--art-padding);
  transition: all var(--art-transition-duration) ease;
  background-size: 100% var(--art-bottom-height);
  background-image: linear-gradient(to top, #000, rgba(0, 0, 0, 0.4), transparent);
  background-repeat: repeat-x;
  background-position: center bottom;
}
.art-video-player .art-bottom .art-controls,
.art-video-player .art-bottom .art-progress {
  transform: translateY(var(--art-bottom-offset));
  transition: transform var(--art-transition-duration) ease;
}
.art-video-player.art-control-show .art-bottom,
.art-video-player.art-hover .art-bottom {
  opacity: 1;
}
.art-video-player.art-control-show .art-bottom .art-controls,
.art-video-player.art-hover .art-bottom .art-controls,
.art-video-player.art-control-show .art-bottom .art-progress,
.art-video-player.art-hover .art-bottom .art-progress {
  transform: translateY(0);
}
.art-bottom .art-progress {
  position: relative;
  z-index: 0;
  cursor: pointer;
  pointer-events: auto;
  padding-top: var(--art-progress-top-gap);
  padding-bottom: var(--art-bottom-gap);
}
.art-bottom .art-progress .art-control-progress {
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  height: var(--art-progress-height);
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner {
  display: flex;
  align-items: center;
  position: relative;
  height: 50%;
  width: 100%;
  transition: height var(--art-transition-duration) ease;
  background-color: var(--art-progress-color);
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-hover {
  position: absolute;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
  width: 0%;
  background-color: var(--art-hover-color);
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-loaded {
  position: absolute;
  inset: 0;
  z-index: 10;
  width: 100%;
  height: 100%;
  width: 0%;
  background-color: var(--art-loaded-color);
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-played {
  position: absolute;
  inset: 0;
  z-index: 20;
  width: 100%;
  height: 100%;
  width: 0%;
  background-color: var(--art-theme);
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-highlight {
  position: absolute;
  inset: 0;
  z-index: 30;
  width: 100%;
  height: 100%;
  pointer-events: none;
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-highlight span {
  position: absolute;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
  right: auto;
  pointer-events: auto;
  width: var(--art-highlight-width) !important;
  transform: translateX(calc(var(--art-highlight-width) / -2));
  background-color: var(--art-highlight-color);
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-indicator {
  display: flex;
  justify-content: center;
  align-items: center;
  position: absolute;
  z-index: 40;
  left: 0;
  border-radius: 50%;
  width: var(--art-indicator-size);
  height: var(--art-indicator-size);
  transform: scale(var(--art-indicator-scale));
  margin-left: calc(var(--art-indicator-size) / -2);
  transition: transform var(--art-transition-duration) ease;
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-indicator .art-icon {
  width: 100%;
  height: 100%;
  pointer-events: none;
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-indicator:hover {
  transform: scale(1.2) !important;
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-indicator:active {
  transform: scale(1) !important;
}
.art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-tip {
  transform-origin: bottom center;
  transform: scale(0.5);
  opacity: 0;
  position: absolute;
  z-index: 50;
  top: -25px;
  left: 0;
  padding: 3px 5px;
  line-height: 1;
  font-size: 12px;
  border-radius: var(--art-border-radius);
  white-space: nowrap;
  background-color: var(--art-tip-background);
  transition: transform var(--art-transition-duration) ease, opacity var(--art-transition-duration) ease;
}
.art-bottom .art-progress .art-control-thumbnails {
  transform-origin: bottom center;
  transform: scale(0.5);
  opacity: 0;
  position: absolute;
  bottom: calc(var(--art-bottom-gap) + 10px);
  left: 0;
  border-radius: var(--art-border-radius);
  pointer-events: none;
  background-color: var(--art-widget-background);
  transition: transform var(--art-transition-duration) ease, opacity var(--art-transition-duration) ease;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.2), 0 1px 2px -1px rgba(0, 0, 0, 0.2);
}
.art-bottom .art-progress:hover .art-control-progress .art-control-progress-inner {
  height: 100%;
}
.art-bottom:hover .art-progress .art-control-progress .art-control-progress-inner .art-progress-indicator {
  transform: scale(1);
}
.art-progress-hover .art-bottom .art-progress .art-control-progress .art-control-progress-inner .art-progress-tip,
.art-progress-hover .art-bottom .art-progress .art-control-thumbnails {
  transform: scale(1);
  opacity: 1;
}
.art-video-player .art-controls {
  position: relative;
  z-index: 10;
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--art-control-height);
}
.art-video-player .art-controls .art-controls-left,
.art-video-player .art-controls .art-controls-right {
  display: flex;
  height: 100%;
}
.art-video-player .art-controls .art-controls-center {
  display: none;
  justify-content: center;
  align-items: center;
  flex: 1;
  height: 100%;
  padding: 0 10px;
}
.art-video-player .art-controls .art-controls-right {
  justify-content: flex-end;
}
.art-video-player .art-controls .art-control {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-shrink: 0;
  cursor: pointer;
  white-space: nowrap;
  opacity: var(--art-control-opacity);
  min-height: var(--art-control-height);
  min-width: var(--art-control-height);
  transition: opacity var(--art-transition-duration) ease;
}
.art-video-player .art-controls .art-control .art-icon {
  height: var(--art-control-icon-size);
  width: var(--art-control-icon-size);
  transform: scale(var(--art-control-icon-scale));
  transition: transform var(--art-transition-duration) ease;
}
.art-video-player .art-controls .art-control .art-icon:active {
  transform: scale(calc(var(--art-control-icon-scale) * 0.8));
}
.art-video-player .art-controls .art-control:hover {
  opacity: 1;
}
.art-control-volume {
  position: relative;
}
.art-control-volume .art-volume-panel {
  display: flex;
  justify-content: center;
  align-items: center;
  position: absolute;
  left: 0;
  right: 0;
  padding: 0 5px;
  font-size: 12px;
  text-align: center;
  cursor: default;
  opacity: 0;
  transform: translateY(10px);
  pointer-events: none;
  bottom: var(--art-control-height);
  width: var(--art-control-height);
  height: var(--art-volume-height);
  transition: all var(--art-transition-duration) ease;
}
.art-control-volume .art-volume-panel .art-volume-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  height: 100%;
  width: 100%;
  padding: 10px 0 12px;
  border-radius: var(--art-border-radius);
  background-color: var(--art-widget-background);
}
.art-control-volume .art-volume-panel .art-volume-inner .art-volume-slider {
  flex: 1;
  width: 100%;
  display: flex;
  cursor: pointer;
  position: relative;
  justify-content: center;
}
.art-control-volume .art-volume-panel .art-volume-inner .art-volume-slider .art-volume-handle {
  position: relative;
  display: flex;
  justify-content: center;
  width: 2px;
  border-radius: var(--art-border-radius);
  overflow: hidden;
  background-color: rgba(255, 255, 255, 0.25);
}
.art-control-volume .art-volume-panel .art-volume-inner .art-volume-slider .art-volume-handle .art-volume-loaded {
  position: absolute;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
  background-color: var(--art-theme);
}
.art-control-volume .art-volume-panel .art-volume-inner .art-volume-slider .art-volume-indicator {
  position: absolute;
  width: var(--art-volume-handle-size);
  height: var(--art-volume-handle-size);
  margin-top: calc(var(--art-volume-handle-size) / -2);
  flex-shrink: 0;
  transform: scale(1);
  border-radius: 100%;
  background-color: var(--art-theme);
  transition: transform var(--art-transition-duration) ease;
}
.art-control-volume .art-volume-panel .art-volume-inner .art-volume-slider:active .art-volume-indicator {
  transform: scale(0.9);
}
.art-control-volume:hover .art-volume-panel {
  opacity: 1;
  transform: translateY(0);
  pointer-events: auto;
}
.art-video-player .art-notice {
  display: none;
  position: absolute;
  inset: 0;
  z-index: 80;
  width: 100%;
  height: 100%;
  height: auto;
  bottom: auto;
  padding: var(--art-padding);
  pointer-events: none;
}
.art-video-player .art-notice .art-notice-inner {
  display: inline-flex;
  padding: 5px;
  line-height: 1;
  border-radius: var(--art-border-radius);
  background-color: var(--art-tip-background);
}
.art-video-player.art-notice-show .art-notice {
  display: flex;
}
.art-video-player .art-contextmenus {
  display: none;
  flex-direction: column;
  position: absolute;
  z-index: 120;
  padding: 5px 0;
  border-radius: var(--art-border-radius);
  font-size: 12px;
  background-color: var(--art-widget-background);
  min-width: var(--art-contextmenus-min-width);
}
.art-video-player .art-contextmenus .art-contextmenu {
  cursor: pointer;
  display: flex;
  padding: 10px 15px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}
.art-video-player .art-contextmenus .art-contextmenu span {
  padding: 0 8px;
}
.art-video-player .art-contextmenus .art-contextmenu span:hover,
.art-video-player .art-contextmenus .art-contextmenu span.art-current {
  color: var(--art-theme);
}
.art-video-player .art-contextmenus .art-contextmenu:hover {
  background-color: rgba(255, 255, 255, 0.1);
}
.art-video-player .art-contextmenus .art-contextmenu:last-child {
  border-bottom: none;
}
.art-video-player.art-contextmenu-show .art-contextmenus {
  display: flex;
}
.art-video-player .art-settings {
  display: none;
  flex-direction: column;
  position: absolute;
  z-index: 90;
  left: auto;
  overflow-y: auto;
  overflow-x: hidden;
  border-radius: var(--art-border-radius);
  max-height: var(--art-settings-max-height);
  right: var(--art-padding);
  bottom: var(--art-control-height);
  transition: all var(--art-transition-duration) ease;
  background-color: var(--art-widget-background);
}
.art-video-player .art-settings .art-setting-panel {
  display: none;
  flex-direction: column;
}
.art-video-player .art-settings .art-setting-panel.art-current {
  display: flex;
}
.art-video-player .art-settings .art-setting-panel .art-setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 5px;
  cursor: pointer;
  overflow: hidden;
  transition: background-color var(--art-transition-duration) ease;
}
.art-video-player .art-settings .art-setting-panel .art-setting-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
}
.art-video-player .art-settings .art-setting-panel .art-setting-item.art-current {
  color: var(--art-theme);
}
.art-video-player .art-settings .art-setting-panel .art-setting-item .art-icon-check {
  visibility: hidden;
  height: 15px;
}
.art-video-player .art-settings .art-setting-panel .art-setting-item.art-current .art-icon-check {
  visibility: visible;
}
.art-video-player .art-settings .art-setting-panel .art-setting-item .art-setting-item-left {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-shrink: 0;
  gap: 5px;
}
.art-video-player .art-settings .art-setting-panel .art-setting-item .art-setting-item-left .art-setting-item-left-icon {
  display: flex;
  justify-content: center;
  align-items: center;
  height: var(--art-settings-icon-size);
  width: var(--art-settings-icon-size);
}
.art-video-player .art-settings .art-setting-panel .art-setting-item .art-setting-item-right {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 5px;
  font-size: 12px;
}
.art-video-player .art-settings .art-setting-panel .art-setting-item .art-setting-item-right .art-setting-item-right-tooltip {
  white-space: nowrap;
  color: rgba(255, 255, 255, 0.5);
}
.art-video-player .art-settings .art-setting-panel .art-setting-item .art-setting-item-right .art-setting-item-right-icon {
  display: flex;
  justify-content: center;
  align-items: center;
  min-width: 32px;
  height: 24px;
}
.art-video-player .art-settings .art-setting-panel .art-setting-item .art-setting-item-right .art-setting-range {
  height: 3px;
  width: 80px;
  outline: none;
  appearance: none;
  background-color: rgba(255, 255, 255, 0.2);
}
.art-video-player .art-settings .art-setting-panel .art-setting-item-back {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}
.art-video-player.art-setting-show .art-settings {
  display: flex;
}
.art-video-player .art-info {
  display: none;
  position: absolute;
  left: var(--art-padding);
  top: var(--art-padding);
  z-index: 100;
  padding: 10px;
  font-size: 12px;
  border-radius: var(--art-border-radius);
  background-color: var(--art-widget-background);
}
.art-video-player .art-info .art-info-panel {
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.art-video-player .art-info .art-info-panel .art-info-item {
  display: flex;
  align-items: center;
  gap: 5px;
}
.art-video-player .art-info .art-info-panel .art-info-item .art-info-title {
  width: 100px;
  text-align: right;
}
.art-video-player .art-info .art-info-panel .art-info-item .art-info-content {
  width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  user-select: all;
}
.art-video-player .art-info .art-info-close {
  position: absolute;
  top: 5px;
  right: 5px;
  cursor: pointer;
}
.art-video-player.art-info-show .art-info {
  display: flex;
}
.art-hide-cursor * {
  cursor: none !important;
}
.art-video-player[data-aspect-ratio] {
  overflow: hidden;
}
.art-video-player[data-aspect-ratio] .art-video {
  object-fit: fill;
  box-sizing: content-box;
}
.art-fullscreen {
  --art-progress-height: 8px;
  --art-indicator-size: 20px;
  --art-control-height: 60px;
  --art-control-icon-scale: 1.3;
}
.art-fullscreen-web {
  --art-progress-height: 8px;
  --art-indicator-size: 20px;
  --art-control-height: 60px;
  --art-control-icon-scale: 1.3;
  position: fixed;
  inset: 0;
  z-index: var(--art-fullscreen-web-index);
  width: 100%;
  height: 100%;
}
.art-mini-popup {
  position: fixed;
  z-index: 9999;
  width: 320px;
  height: 180px;
  background: #000;
  border-radius: var(--art-border-radius);
  cursor: move;
  user-select: none;
  overflow: hidden;
  transition: opacity 0.2s ease;
  box-shadow: 0 0 5px rgba(0, 0, 0, 0.5);
}
.art-mini-popup svg {
  fill: #fff;
}
.art-mini-popup .art-video {
  pointer-events: none;
}
.art-mini-popup .art-mini-close {
  position: absolute;
  z-index: 20;
  right: 10px;
  top: 10px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s ease;
}
.art-mini-popup .art-mini-state {
  position: absolute;
  inset: 0;
  z-index: 30;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.2s ease;
  background-color: rgba(0, 0, 0, 0.25);
}
.art-mini-popup .art-mini-state .art-icon {
  opacity: 0.75;
  cursor: pointer;
  transform: scale(3);
  pointer-events: auto;
  transition: transform 0.2s ease;
}
.art-mini-popup .art-mini-state .art-icon:active {
  transform: scale(2.5);
}
.art-mini-popup.art-mini-dragging {
  opacity: 0.9;
}
.art-mini-popup:hover .art-mini-close,
.art-mini-popup:hover .art-mini-state {
  opacity: 1;
}
.art-video-player[data-flip='horizontal'] .art-video {
  transform: scaleX(-1);
}
.art-video-player[data-flip='vertical'] .art-video {
  transform: scaleY(-1);
}
.art-video-player .art-layer-lock {
  display: none;
  justify-content: center;
  align-items: center;
  position: absolute;
  top: 50%;
  border-radius: 50%;
  transform: translateY(-50%);
  height: var(--art-lock-size);
  width: var(--art-lock-size);
  left: var(--art-padding);
  background-color: var(--art-tip-background);
}
.art-video-player .art-layer-auto-playback {
  display: none;
  gap: 10px;
  align-items: center;
  position: absolute;
  border-radius: var(--art-border-radius);
  padding: 10px;
  line-height: 1;
  left: var(--art-padding);
  bottom: calc(var(--art-control-height) + var(--art-bottom-gap) + 10px);
  background-color: var(--art-widget-background);
}
.art-video-player .art-layer-auto-playback .art-auto-playback-close {
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
}
.art-video-player .art-layer-auto-playback .art-auto-playback-close svg {
  width: 15px;
  height: 15px;
  fill: var(--art-theme);
}
.art-video-player .art-layer-auto-playback .art-auto-playback-jump {
  color: var(--art-theme);
  cursor: pointer;
}
.art-video-player.art-lock .art-subtitle {
  bottom: var(--art-subtitle-bottom) !important;
}
.art-video-player.art-mini-progress-bar .art-bottom,
.art-video-player.art-lock .art-bottom {
  opacity: 1;
  padding: 0;
  background-image: none;
}
.art-video-player.art-mini-progress-bar .art-bottom .art-controls,
.art-video-player.art-lock .art-bottom .art-controls,
.art-video-player.art-mini-progress-bar .art-bottom .art-progress,
.art-video-player.art-lock .art-bottom .art-progress {
  transform: translateY(calc(var(--art-control-height) + var(--art-bottom-gap) + var(--art-progress-height) / 4));
}
.art-video-player.art-mini-progress-bar .art-bottom .art-progress-indicator,
.art-video-player.art-lock .art-bottom .art-progress-indicator {
  display: none !important;
}
.art-video-player.art-control-show .art-layer-lock {
  display: flex;
}
.art-control-selector {
  position: relative;
  display: flex;
  justify-content: center;
}
.art-control-selector .art-selector-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  position: absolute;
  border-radius: var(--art-border-radius);
  overflow-y: auto;
  overflow-x: hidden;
  opacity: 0;
  transform: translateY(10px);
  pointer-events: none;
  bottom: var(--art-control-height);
  max-height: var(--art-selector-max-height);
  background-color: var(--art-widget-background);
  transition: all var(--art-transition-duration) ease;
}
.art-control-selector .art-selector-list .art-selector-item {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  padding: 10px 15px;
  flex-shrink: 0;
  line-height: 1;
}
.art-control-selector .art-selector-list .art-selector-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
}
.art-control-selector .art-selector-list .art-selector-item:hover,
.art-control-selector .art-selector-list .art-selector-item.art-current {
  color: var(--art-theme);
}
.art-control-selector:hover .art-selector-list {
  opacity: 1;
  transform: translateY(0);
  pointer-events: auto;
}
.art-video-player {
  /*! Hint.css - v2.7.0 - 2021-10-01
    * https://kushagra.dev/lab/hint/
    * Copyright (c) 2021 Kushagra Gour */
  /*-------------------------------------*\\
        HINT.css - A CSS tooltip library
    \\*-------------------------------------*/
  /**
    * HINT.css is a tooltip library made in pure CSS.
    *
    * Source: https://github.com/chinchang/hint.css
    * Demo: http://kushagragour.in/lab/hint/
    *
    */
  /**
    * source: hint-core.scss
    *
    * Defines the basic styling for the tooltip.
    * Each tooltip is made of 2 parts:
    * 	1) body (:after)
    * 	2) arrow (:before)
    *
    * Classes added:
    * 	1) hint
    */
  /**
    * source: hint-position.scss
    *
    * Defines the positoning logic for the tooltips.
    *
    * Classes added:
    * 	1) hint--top
    * 	2) hint--bottom
    * 	3) hint--left
    * 	4) hint--right
    */
  /**
    * set default color for tooltip arrows
    */
  /**
    * top tooltip
    */
  /**
    * bottom tooltip
    */
  /**
    * right tooltip
    */
  /**
    * left tooltip
    */
  /**
    * top-left tooltip
    */
  /**
    * top-right tooltip
    */
  /**
    * bottom-left tooltip
    */
  /**
    * bottom-right tooltip
    */
  /**
    * source: hint-sizes.scss
    *
    * Defines width restricted tooltips that can span
    * across multiple lines.
    *
    * Classes added:
    * 	1) hint--small
    * 	2) hint--medium
    * 	3) hint--large
    *
    */
  /**
    * source: hint-theme.scss
    *
    * Defines basic theme for tooltips.
    *
    */
  /**
    * source: hint-color-types.scss
    *
    * Contains tooltips of various types based on color differences.
    *
    * Classes added:
    * 	1) hint--error
    * 	2) hint--warning
    * 	3) hint--info
    * 	4) hint--success
    *
    */
  /**
    * Error
    */
  /**
    * Warning
    */
  /**
    * Info
    */
  /**
    * Success
    */
  /**
    * source: hint-always.scss
    *
    * Defines a persisted tooltip which shows always.
    *
    * Classes added:
    * 	1) hint--always
    *
    */
  /**
    * source: hint-rounded.scss
    *
    * Defines rounded corner tooltips.
    *
    * Classes added:
    * 	1) hint--rounded
    *
    */
  /**
    * source: hint-effects.scss
    *
    * Defines various transition effects for the tooltips.
    *
    * Classes added:
    * 	1) hint--no-animate
    * 	2) hint--bounce
    *
    */
}
.art-video-player [class*='hint--'] {
  position: relative;
  display: inline-block;
  font-style: normal;
  /**
        * tooltip arrow
        */
  /**
        * tooltip body
        */
}
.art-video-player [class*='hint--']:before,
.art-video-player [class*='hint--']:after {
  position: absolute;
  -webkit-transform: translate3d(0, 0, 0);
  -moz-transform: translate3d(0, 0, 0);
  transform: translate3d(0, 0, 0);
  visibility: hidden;
  opacity: 0;
  z-index: 1000000;
  pointer-events: none;
  -webkit-transition: 0.3s ease;
  -moz-transition: 0.3s ease;
  transition: 0.3s ease;
  -webkit-transition-delay: 0ms;
  -moz-transition-delay: 0ms;
  transition-delay: 0ms;
}
.art-video-player [class*='hint--']:hover:before,
.art-video-player [class*='hint--']:hover:after {
  visibility: visible;
  opacity: 1;
}
.art-video-player [class*='hint--']:hover:before,
.art-video-player [class*='hint--']:hover:after {
  -webkit-transition-delay: 100ms;
  -moz-transition-delay: 100ms;
  transition-delay: 100ms;
}
.art-video-player [class*='hint--']:before {
  content: '';
  position: absolute;
  background: transparent;
  border: 6px solid transparent;
  z-index: 1000001;
}
.art-video-player [class*='hint--']:after {
  background: #000000;
  color: white;
  padding: 8px 10px;
  font-size: 12px;
  font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
  line-height: 12px;
  white-space: nowrap;
}
.art-video-player [class*='hint--'][aria-label]:after {
  content: attr(aria-label);
}
.art-video-player [class*='hint--'][data-hint]:after {
  content: attr(data-hint);
}
.art-video-player [aria-label='']:before,
.art-video-player [aria-label='']:after,
.art-video-player [data-hint='']:before,
.art-video-player [data-hint='']:after {
  display: none !important;
}
.art-video-player .hint--top-left:before {
  border-top-color: #000000;
}
.art-video-player .hint--top-right:before {
  border-top-color: #000000;
}
.art-video-player .hint--top:before {
  border-top-color: #000000;
}
.art-video-player .hint--bottom-left:before {
  border-bottom-color: #000000;
}
.art-video-player .hint--bottom-right:before {
  border-bottom-color: #000000;
}
.art-video-player .hint--bottom:before {
  border-bottom-color: #000000;
}
.art-video-player .hint--left:before {
  border-left-color: #000000;
}
.art-video-player .hint--right:before {
  border-right-color: #000000;
}
.art-video-player .hint--top:before {
  margin-bottom: -11px;
}
.art-video-player .hint--top:before,
.art-video-player .hint--top:after {
  bottom: 100%;
  left: 50%;
}
.art-video-player .hint--top:before {
  left: calc(50% - 6px);
}
.art-video-player .hint--top:after {
  -webkit-transform: translateX(-50%);
  -moz-transform: translateX(-50%);
  transform: translateX(-50%);
}
.art-video-player .hint--top:hover:before {
  -webkit-transform: translateY(-8px);
  -moz-transform: translateY(-8px);
  transform: translateY(-8px);
}
.art-video-player .hint--top:hover:after {
  -webkit-transform: translateX(-50%) translateY(-8px);
  -moz-transform: translateX(-50%) translateY(-8px);
  transform: translateX(-50%) translateY(-8px);
}
.art-video-player .hint--bottom:before {
  margin-top: -11px;
}
.art-video-player .hint--bottom:before,
.art-video-player .hint--bottom:after {
  top: 100%;
  left: 50%;
}
.art-video-player .hint--bottom:before {
  left: calc(50% - 6px);
}
.art-video-player .hint--bottom:after {
  -webkit-transform: translateX(-50%);
  -moz-transform: translateX(-50%);
  transform: translateX(-50%);
}
.art-video-player .hint--bottom:hover:before {
  -webkit-transform: translateY(8px);
  -moz-transform: translateY(8px);
  transform: translateY(8px);
}
.art-video-player .hint--bottom:hover:after {
  -webkit-transform: translateX(-50%) translateY(8px);
  -moz-transform: translateX(-50%) translateY(8px);
  transform: translateX(-50%) translateY(8px);
}
.art-video-player .hint--right:before {
  margin-left: -11px;
  margin-bottom: -6px;
}
.art-video-player .hint--right:after {
  margin-bottom: -14px;
}
.art-video-player .hint--right:before,
.art-video-player .hint--right:after {
  left: 100%;
  bottom: 50%;
}
.art-video-player .hint--right:hover:before {
  -webkit-transform: translateX(8px);
  -moz-transform: translateX(8px);
  transform: translateX(8px);
}
.art-video-player .hint--right:hover:after {
  -webkit-transform: translateX(8px);
  -moz-transform: translateX(8px);
  transform: translateX(8px);
}
.art-video-player .hint--left:before {
  margin-right: -11px;
  margin-bottom: -6px;
}
.art-video-player .hint--left:after {
  margin-bottom: -14px;
}
.art-video-player .hint--left:before,
.art-video-player .hint--left:after {
  right: 100%;
  bottom: 50%;
}
.art-video-player .hint--left:hover:before {
  -webkit-transform: translateX(-8px);
  -moz-transform: translateX(-8px);
  transform: translateX(-8px);
}
.art-video-player .hint--left:hover:after {
  -webkit-transform: translateX(-8px);
  -moz-transform: translateX(-8px);
  transform: translateX(-8px);
}
.art-video-player .hint--top-left:before {
  margin-bottom: -11px;
}
.art-video-player .hint--top-left:before,
.art-video-player .hint--top-left:after {
  bottom: 100%;
  left: 50%;
}
.art-video-player .hint--top-left:before {
  left: calc(50% - 6px);
}
.art-video-player .hint--top-left:after {
  -webkit-transform: translateX(-100%);
  -moz-transform: translateX(-100%);
  transform: translateX(-100%);
}
.art-video-player .hint--top-left:after {
  margin-left: 12px;
}
.art-video-player .hint--top-left:hover:before {
  -webkit-transform: translateY(-8px);
  -moz-transform: translateY(-8px);
  transform: translateY(-8px);
}
.art-video-player .hint--top-left:hover:after {
  -webkit-transform: translateX(-100%) translateY(-8px);
  -moz-transform: translateX(-100%) translateY(-8px);
  transform: translateX(-100%) translateY(-8px);
}
.art-video-player .hint--top-right:before {
  margin-bottom: -11px;
}
.art-video-player .hint--top-right:before,
.art-video-player .hint--top-right:after {
  bottom: 100%;
  left: 50%;
}
.art-video-player .hint--top-right:before {
  left: calc(50% - 6px);
}
.art-video-player .hint--top-right:after {
  -webkit-transform: translateX(0);
  -moz-transform: translateX(0);
  transform: translateX(0);
}
.art-video-player .hint--top-right:after {
  margin-left: -12px;
}
.art-video-player .hint--top-right:hover:before {
  -webkit-transform: translateY(-8px);
  -moz-transform: translateY(-8px);
  transform: translateY(-8px);
}
.art-video-player .hint--top-right:hover:after {
  -webkit-transform: translateY(-8px);
  -moz-transform: translateY(-8px);
  transform: translateY(-8px);
}
.art-video-player .hint--bottom-left:before {
  margin-top: -11px;
}
.art-video-player .hint--bottom-left:before,
.art-video-player .hint--bottom-left:after {
  top: 100%;
  left: 50%;
}
.art-video-player .hint--bottom-left:before {
  left: calc(50% - 6px);
}
.art-video-player .hint--bottom-left:after {
  -webkit-transform: translateX(-100%);
  -moz-transform: translateX(-100%);
  transform: translateX(-100%);
}
.art-video-player .hint--bottom-left:after {
  margin-left: 12px;
}
.art-video-player .hint--bottom-left:hover:before {
  -webkit-transform: translateY(8px);
  -moz-transform: translateY(8px);
  transform: translateY(8px);
}
.art-video-player .hint--bottom-left:hover:after {
  -webkit-transform: translateX(-100%) translateY(8px);
  -moz-transform: translateX(-100%) translateY(8px);
  transform: translateX(-100%) translateY(8px);
}
.art-video-player .hint--bottom-right:before {
  margin-top: -11px;
}
.art-video-player .hint--bottom-right:before,
.art-video-player .hint--bottom-right:after {
  top: 100%;
  left: 50%;
}
.art-video-player .hint--bottom-right:before {
  left: calc(50% - 6px);
}
.art-video-player .hint--bottom-right:after {
  -webkit-transform: translateX(0);
  -moz-transform: translateX(0);
  transform: translateX(0);
}
.art-video-player .hint--bottom-right:after {
  margin-left: -12px;
}
.art-video-player .hint--bottom-right:hover:before {
  -webkit-transform: translateY(8px);
  -moz-transform: translateY(8px);
  transform: translateY(8px);
}
.art-video-player .hint--bottom-right:hover:after {
  -webkit-transform: translateY(8px);
  -moz-transform: translateY(8px);
  transform: translateY(8px);
}
.art-video-player .hint--small:after,
.art-video-player .hint--medium:after,
.art-video-player .hint--large:after {
  white-space: normal;
  line-height: 1.4em;
  word-wrap: break-word;
}
.art-video-player .hint--small:after {
  width: 80px;
}
.art-video-player .hint--medium:after {
  width: 150px;
}
.art-video-player .hint--large:after {
  width: 300px;
}
.art-video-player [class*='hint--'] {
  /**
        * tooltip body
        */
}
.art-video-player [class*='hint--']:after {
  text-shadow: 0 -1px 0px black;
  box-shadow: 4px 4px 8px rgba(0, 0, 0, 0.3);
}
.art-video-player .hint--error:after {
  background-color: #b34e4d;
  text-shadow: 0 -1px 0px #592726;
}
.art-video-player .hint--error.hint--top-left:before {
  border-top-color: #b34e4d;
}
.art-video-player .hint--error.hint--top-right:before {
  border-top-color: #b34e4d;
}
.art-video-player .hint--error.hint--top:before {
  border-top-color: #b34e4d;
}
.art-video-player .hint--error.hint--bottom-left:before {
  border-bottom-color: #b34e4d;
}
.art-video-player .hint--error.hint--bottom-right:before {
  border-bottom-color: #b34e4d;
}
.art-video-player .hint--error.hint--bottom:before {
  border-bottom-color: #b34e4d;
}
.art-video-player .hint--error.hint--left:before {
  border-left-color: #b34e4d;
}
.art-video-player .hint--error.hint--right:before {
  border-right-color: #b34e4d;
}
.art-video-player .hint--warning:after {
  background-color: #c09854;
  text-shadow: 0 -1px 0px #6c5328;
}
.art-video-player .hint--warning.hint--top-left:before {
  border-top-color: #c09854;
}
.art-video-player .hint--warning.hint--top-right:before {
  border-top-color: #c09854;
}
.art-video-player .hint--warning.hint--top:before {
  border-top-color: #c09854;
}
.art-video-player .hint--warning.hint--bottom-left:before {
  border-bottom-color: #c09854;
}
.art-video-player .hint--warning.hint--bottom-right:before {
  border-bottom-color: #c09854;
}
.art-video-player .hint--warning.hint--bottom:before {
  border-bottom-color: #c09854;
}
.art-video-player .hint--warning.hint--left:before {
  border-left-color: #c09854;
}
.art-video-player .hint--warning.hint--right:before {
  border-right-color: #c09854;
}
.art-video-player .hint--info:after {
  background-color: #3986ac;
  text-shadow: 0 -1px 0px #1a3c4d;
}
.art-video-player .hint--info.hint--top-left:before {
  border-top-color: #3986ac;
}
.art-video-player .hint--info.hint--top-right:before {
  border-top-color: #3986ac;
}
.art-video-player .hint--info.hint--top:before {
  border-top-color: #3986ac;
}
.art-video-player .hint--info.hint--bottom-left:before {
  border-bottom-color: #3986ac;
}
.art-video-player .hint--info.hint--bottom-right:before {
  border-bottom-color: #3986ac;
}
.art-video-player .hint--info.hint--bottom:before {
  border-bottom-color: #3986ac;
}
.art-video-player .hint--info.hint--left:before {
  border-left-color: #3986ac;
}
.art-video-player .hint--info.hint--right:before {
  border-right-color: #3986ac;
}
.art-video-player .hint--success:after {
  background-color: #458746;
  text-shadow: 0 -1px 0px #1a321a;
}
.art-video-player .hint--success.hint--top-left:before {
  border-top-color: #458746;
}
.art-video-player .hint--success.hint--top-right:before {
  border-top-color: #458746;
}
.art-video-player .hint--success.hint--top:before {
  border-top-color: #458746;
}
.art-video-player .hint--success.hint--bottom-left:before {
  border-bottom-color: #458746;
}
.art-video-player .hint--success.hint--bottom-right:before {
  border-bottom-color: #458746;
}
.art-video-player .hint--success.hint--bottom:before {
  border-bottom-color: #458746;
}
.art-video-player .hint--success.hint--left:before {
  border-left-color: #458746;
}
.art-video-player .hint--success.hint--right:before {
  border-right-color: #458746;
}
.art-video-player .hint--always:after,
.art-video-player .hint--always:before {
  opacity: 1;
  visibility: visible;
}
.art-video-player .hint--always.hint--top:before {
  -webkit-transform: translateY(-8px);
  -moz-transform: translateY(-8px);
  transform: translateY(-8px);
}
.art-video-player .hint--always.hint--top:after {
  -webkit-transform: translateX(-50%) translateY(-8px);
  -moz-transform: translateX(-50%) translateY(-8px);
  transform: translateX(-50%) translateY(-8px);
}
.art-video-player .hint--always.hint--top-left:before {
  -webkit-transform: translateY(-8px);
  -moz-transform: translateY(-8px);
  transform: translateY(-8px);
}
.art-video-player .hint--always.hint--top-left:after {
  -webkit-transform: translateX(-100%) translateY(-8px);
  -moz-transform: translateX(-100%) translateY(-8px);
  transform: translateX(-100%) translateY(-8px);
}
.art-video-player .hint--always.hint--top-right:before {
  -webkit-transform: translateY(-8px);
  -moz-transform: translateY(-8px);
  transform: translateY(-8px);
}
.art-video-player .hint--always.hint--top-right:after {
  -webkit-transform: translateY(-8px);
  -moz-transform: translateY(-8px);
  transform: translateY(-8px);
}
.art-video-player .hint--always.hint--bottom:before {
  -webkit-transform: translateY(8px);
  -moz-transform: translateY(8px);
  transform: translateY(8px);
}
.art-video-player .hint--always.hint--bottom:after {
  -webkit-transform: translateX(-50%) translateY(8px);
  -moz-transform: translateX(-50%) translateY(8px);
  transform: translateX(-50%) translateY(8px);
}
.art-video-player .hint--always.hint--bottom-left:before {
  -webkit-transform: translateY(8px);
  -moz-transform: translateY(8px);
  transform: translateY(8px);
}
.art-video-player .hint--always.hint--bottom-left:after {
  -webkit-transform: translateX(-100%) translateY(8px);
  -moz-transform: translateX(-100%) translateY(8px);
  transform: translateX(-100%) translateY(8px);
}
.art-video-player .hint--always.hint--bottom-right:before {
  -webkit-transform: translateY(8px);
  -moz-transform: translateY(8px);
  transform: translateY(8px);
}
.art-video-player .hint--always.hint--bottom-right:after {
  -webkit-transform: translateY(8px);
  -moz-transform: translateY(8px);
  transform: translateY(8px);
}
.art-video-player .hint--always.hint--left:before {
  -webkit-transform: translateX(-8px);
  -moz-transform: translateX(-8px);
  transform: translateX(-8px);
}
.art-video-player .hint--always.hint--left:after {
  -webkit-transform: translateX(-8px);
  -moz-transform: translateX(-8px);
  transform: translateX(-8px);
}
.art-video-player .hint--always.hint--right:before {
  -webkit-transform: translateX(8px);
  -moz-transform: translateX(8px);
  transform: translateX(8px);
}
.art-video-player .hint--always.hint--right:after {
  -webkit-transform: translateX(8px);
  -moz-transform: translateX(8px);
  transform: translateX(8px);
}
.art-video-player .hint--rounded:after {
  border-radius: 4px;
}
.art-video-player .hint--no-animate:before,
.art-video-player .hint--no-animate:after {
  -webkit-transition-duration: 0ms;
  -moz-transition-duration: 0ms;
  transition-duration: 0ms;
}
.art-video-player .hint--bounce:before,
.art-video-player .hint--bounce:after {
  -webkit-transition: opacity 0.3s ease, visibility 0.3s ease, -webkit-transform 0.3s cubic-bezier(0.71, 1.7, 0.77, 1.24);
  -moz-transition: opacity 0.3s ease, visibility 0.3s ease, -moz-transform 0.3s cubic-bezier(0.71, 1.7, 0.77, 1.24);
  transition: opacity 0.3s ease, visibility 0.3s ease, transform 0.3s cubic-bezier(0.71, 1.7, 0.77, 1.24);
}
.art-video-player .hint--no-shadow:before,
.art-video-player .hint--no-shadow:after {
  text-shadow: initial;
  box-shadow: initial;
}
.art-video-player .hint--no-arrow:before {
  display: none;
}
.art-video-player.art-mobile {
  --art-bottom-gap: 10px;
  --art-control-height: 38px;
  --art-control-icon-scale: 1;
  --art-state-size: 60px;
  --art-settings-max-height: 180px;
  --art-selector-max-height: 180px;
  --art-indicator-scale: 1;
  --art-control-opacity: 1;
}
.art-video-player.art-mobile .art-controls-left {
  margin-left: calc(var(--art-padding) / -1);
}
.art-video-player.art-mobile .art-controls-right {
  margin-right: calc(var(--art-padding) / -1);
}
`;class mo extends it{constructor(t){super(t),this.name="subtitle",this.option=null,this.destroyEvent=()=>null,this.init(t.option.subtitle);let e=!1;t.on("video:timeupdate",()=>{if(!this.url)return;const i=this.art.template.$video.webkitDisplayingFullscreen;typeof i=="boolean"&&i!==e&&(e=i,this.createTrack(i?"subtitles":"metadata",this.url))})}get url(){return this.art.template.$track.src}set url(t){this.switch(t)}get textTrack(){return this.art.template.$video?.textTracks?.[0]}get activeCues(){return this.textTrack?Array.from(this.textTrack.activeCues):[]}get cues(){return this.textTrack?Array.from(this.textTrack.cues):[]}style(t,e){const{$subtitle:i}=this.art.template;return typeof t=="object"?Mt(i,t):h(i,t,e)}update(){const{option:{subtitle:t},template:{$subtitle:e}}=this.art;e.innerHTML="",this.activeCues.length&&(this.art.emit("subtitleBeforeUpdate",this.activeCues),e.innerHTML=this.activeCues.map((i,o)=>i.text.split(/\r?\n/).filter(r=>r.trim()).map(r=>`<div class="art-subtitle-line" data-group="${o}">
                                ${t.escape?$e(r):r}
                            </div>`).join("")).join(""),this.art.emit("subtitleAfterUpdate",this.activeCues))}async switch(t,e={}){const{i18n:i,notice:o,option:r}=this.art,a={...r.subtitle,...e,url:t},s=await this.init(a);return e.name&&(o.show=`${i.get("Switch Subtitle")}: ${e.name}`),s}createTrack(t,e){const{template:i,proxy:o,option:r}=this.art,{$video:a,$track:s}=i,c=D("track");c.default=!0,c.kind=t,c.src=e,c.label=r.subtitle.name||"Artplayer",c.track.mode="hidden",c.onload=()=>{this.art.emit("subtitleLoad",this.cues,this.option)},this.art.events.remove(this.destroyEvent),s.onload=null,Ct(s),k(a,c),i.$track=c,this.destroyEvent=o(this.textTrack,"cuechange",()=>this.update())}async init(t){const{notice:e,template:{$subtitle:i}}=this.art;if(!this.textTrack)return null;if(pt(t,Ot.subtitle),!!t.url)return this.option=t,this.style(t.style),fetch(t.url).then(o=>o.arrayBuffer()).then(o=>{const a=new TextDecoder(t.encoding).decode(o);switch(t.type||ft(t.url)){case"srt":{const s=Ce(a),c=t.onVttLoad(s);return xt(c)}case"ass":{const s=Me(a),c=t.onVttLoad(s);return xt(c)}case"vtt":{const s=t.onVttLoad(a);return xt(s)}default:return t.url}}).then(o=>(i.innerHTML="",this.url===o||(URL.revokeObjectURL(this.url),this.createTrack("metadata",o)),o)).catch(o=>{throw i.innerHTML="",e.show=o,o})}}class kt{constructor(t){this.art=t;const{option:e,constructor:i}=t;e.container instanceof Element?this.$container=e.container:(this.$container=W(e.container),q(this.$container,`No container element found by ${e.container}`)),q(be(),"The current browser does not support flex layout");const o=this.$container.tagName.toLowerCase();q(o==="div",`Unsupported container element type, only support 'div' but got '${o}'`),q(i.instances.every(r=>r.template.$container!==this.$container),"Cannot mount multiple instances on the same dom element"),this.query=this.query.bind(this),this.$container.dataset.artId=t.id,this.init()}static get html(){return`
          <div class="art-video-player art-subtitle-show art-layer-show art-control-show art-mask-show">
            <video class="art-video">
              <track default kind="metadata" src=""></track>
            </video>
            <div class="art-poster"></div>
            <div class="art-subtitle"></div>
            <div class="art-danmuku"></div>
            <div class="art-layers"></div>
            <div class="art-mask">
              <div class="art-state"></div>
            </div>
            <div class="art-bottom">
              <div class="art-progress"></div>
              <div class="art-controls">
                <div class="art-controls-left"></div>
                <div class="art-controls-center"></div>
                <div class="art-controls-right"></div>
              </div>
            </div>
            <div class="art-loading"></div>
            <div class="art-notice">
              <div class="art-notice-inner"></div>
            </div>
            <div class="art-settings"></div>
            <div class="art-info">
              <div class="art-info-panel">
                <div class="art-info-item">
                  <div class="art-info-title">Player version:</div>
                  <div class="art-info-content">${Ht}</div>
                </div>
                <div class="art-info-item">
                  <div class="art-info-title">Video url:</div>
                  <div class="art-info-content" data-video="currentSrc"></div>
                </div>
                <div class="art-info-item">
                  <div class="art-info-title">Video volume:</div>
                  <div class="art-info-content" data-video="volume"></div>
                </div>
                <div class="art-info-item">
                  <div class="art-info-title">Video time:</div>
                  <div class="art-info-content" data-video="currentTime"></div>
                </div>
                <div class="art-info-item">
                  <div class="art-info-title">Video duration:</div>
                  <div class="art-info-content" data-video="duration"></div>
                </div>
                <div class="art-info-item">
                  <div class="art-info-title">Video resolution:</div>
                  <div class="art-info-content">
                    <span data-video="videoWidth"></span> x <span data-video="videoHeight"></span>
                  </div>
                </div>
              </div>
              <div class="art-info-close">[x]</div>
            </div>
            <div class="art-contextmenus"></div>
          </div>
        `}query(t){return W(t,this.$container)}init(){const{option:t}=this.art;if(t.useSSR||(this.$container.innerHTML=kt.html),this.$player=this.query(".art-video-player"),this.$video=this.query(".art-video"),this.$track=this.query("track"),this.$poster=this.query(".art-poster"),this.$subtitle=this.query(".art-subtitle"),this.$danmuku=this.query(".art-danmuku"),this.$bottom=this.query(".art-bottom"),this.$progress=this.query(".art-progress"),this.$controls=this.query(".art-controls"),this.$controlsLeft=this.query(".art-controls-left"),this.$controlsCenter=this.query(".art-controls-center"),this.$controlsRight=this.query(".art-controls-right"),this.$layer=this.query(".art-layers"),this.$loading=this.query(".art-loading"),this.$notice=this.query(".art-notice"),this.$noticeInner=this.query(".art-notice-inner"),this.$mask=this.query(".art-mask"),this.$state=this.query(".art-state"),this.$setting=this.query(".art-settings"),this.$info=this.query(".art-info"),this.$infoPanel=this.query(".art-info-panel"),this.$infoClose=this.query(".art-info-close"),this.$contextmenu=this.query(".art-contextmenus"),t.proxy){const e=t.proxy.call(this.art,this.art);q(e instanceof HTMLVideoElement||e instanceof HTMLCanvasElement,"Function 'option.proxy' needs to return 'HTMLVideoElement' or 'HTMLCanvasElement'"),Nt(e,this.$video),e.className="art-video",this.$video=e}t.backdrop&&S(this.$player,"art-backdrop"),_&&S(this.$player,"art-mobile")}destroy(t){t?this.$container.innerHTML="":S(this.$player,"art-destroy")}}class ee{on(t,e,i){const o=this.e||(this.e={});return(o[t]||(o[t]=[])).push({fn:e,ctx:i}),this}once(t,e,i){const o=this;function r(...a){o.off(t,r),e.apply(i,a)}return r._=e,this.on(t,r,i)}emit(t,...e){const i=((this.e||(this.e={}))[t]||[]).slice();for(let o=0;o<i.length;o+=1)i[o].fn.apply(i[o].ctx,e);return this}off(t,e){const i=this.e||(this.e={}),o=i[t],r=[];if(o&&e)for(let a=0,s=o.length;a<s;a+=1)o[a].fn!==e&&o[a].fn._!==e&&r.push(o[a]);return r.length?i[t]=r:delete i[t],this}}let go=0;const yt=[];class z extends ee{constructor(t,e){if(super(),!Tt)throw new Error("Artplayer can only be used in the browser environment");this.id=++go;const i=St(z.option,t);if(i.container=t.container,this.option=pt(i,Ot),this.isLock=!1,this.isReady=!1,this.isFocus=!1,this.isInput=!1,this.isRotate=!1,this.isDestroy=!1,this.template=new kt(this),this.events=new Vn(this),this.storage=new fo(this),this.icons=new ui(this),this.i18n=new On(this),this.notice=new vi(this),this.player=new eo(this),this.layers=new fi(this),this.controls=new Cn(this),this.contextmenu=new fn(this),this.subtitle=new mo(this),this.info=new hi(this),this.loading=new mi(this),this.hotkey=new Dn(this),this.mask=new gi(this),this.setting=new ho(this),this.plugins=new so(this),typeof e=="function"&&this.on("ready",()=>e.call(this,this)),z.DEBUG){const o=r=>console.log(`[ART.${this.id}] -> ${r}`);o(`Version@${z.version}`);for(let r=0;r<ut.events.length;r++)this.on(`video:${ut.events[r]}`,a=>o(`Event@${a.type}`))}yt.push(this)}static get instances(){return yt}static get version(){return Ht}static get config(){return ut}static get utils(){return sn}static get scheme(){return Ot}static get Emitter(){return ee}static get validator(){return pt}static get kindOf(){return pt.kindOf}static get html(){return kt.html}static get option(){return{id:"",container:"#artplayer",url:"",poster:"",type:"",theme:"#f00",volume:.7,isLive:!1,muted:!1,autoplay:!1,autoSize:!1,autoMini:!1,loop:!1,flip:!1,playbackRate:!1,aspectRatio:!1,screenshot:!1,setting:!1,hotkey:!0,pip:!1,mutex:!0,backdrop:!0,fullscreen:!1,fullscreenWeb:!1,subtitleOffset:!1,miniProgressBar:!1,useSSR:!1,playsInline:!0,lock:!1,gesture:!0,fastForward:!1,autoPlayback:!1,autoOrientation:!1,airplay:!1,proxy:void 0,layers:[],contextmenu:[],controls:[],settings:[],quality:[],highlight:[],plugins:[],thumbnails:{url:"",number:60,column:10,width:0,height:0,scale:1},subtitle:{url:"",type:"",style:{},name:"",escape:!0,encoding:"utf-8",onVttLoad:t=>t},moreVideoAttr:{controls:!1,preload:he?"auto":"metadata"},i18n:{},icons:{},cssVar:{},customType:{},lang:navigator?.language.toLowerCase()}}get proxy(){return this.events.proxy}get query(){return this.template.query}get video(){return this.template.$video}reset(){this.video.removeAttribute("src"),this.video.load()}destroy(t=!0){z.REMOVE_SRC_WHEN_DESTROY&&this.reset(),this.events.destroy(),this.template.destroy(t),yt.splice(yt.indexOf(this),1),this.isDestroy=!0,this.emit("destroy")}}z.STYLE=_e;z.DEBUG=!1;z.CONTEXTMENU=!0;z.NOTICE_TIME=2e3;z.SETTING_WIDTH=250;z.SETTING_ITEM_WIDTH=200;z.SETTING_ITEM_HEIGHT=35;z.RESIZE_TIME=200;z.SCROLL_TIME=200;z.SCROLL_GAP=50;z.AUTO_PLAYBACK_MAX=10;z.AUTO_PLAYBACK_MIN=5;z.AUTO_PLAYBACK_TIMEOUT=3e3;z.RECONNECT_TIME_MAX=5;z.RECONNECT_SLEEP_TIME=1e3;z.CONTROL_HIDE_TIME=3e3;z.DBCLICK_TIME=300;z.DBCLICK_FULLSCREEN=!0;z.MOBILE_DBCLICK_PLAY=!0;z.MOBILE_CLICK_PLAY=!1;z.AUTO_ORIENTATION_TIME=200;z.INFO_LOOP_TIME=1e3;z.FAST_FORWARD_VALUE=3;z.FAST_FORWARD_TIME=1e3;z.TOUCH_MOVE_RATIO=.5;z.VOLUME_STEP=.1;z.SEEK_STEP=5;z.PLAYBACK_RATE=[.5,.75,1,1.25,1.5,2];z.ASPECT_RATIO=["default","4:3","16:9"];z.FLIP=["normal","horizontal","vertical"];z.FULLSCREEN_WEB_IN_BODY=!0;z.LOG_VERSION=!0;z.USE_RAF=!1;z.REMOVE_SRC_WHEN_DESTROY=!0;Tt&&(ye("artplayer-style",_e),setTimeout(()=>{z.LOG_VERSION&&console.log(`%c ArtPlayer %c ${z.version} %c https://artplayer.org`,"color: #fff; background: #5f5f5f","color: #fff; background: #4bc729","")},100));function Ie(n){switch(n){case 1:case 2:case 3:return 0;case 4:return 2;case 5:return 1;default:return 0}}function jt(n){if(typeof n!="string")return[];const t=/<d[^>]*?p="(?<p>[^"]+)"[^>]*>(?<text>.*?)<\/d>/gs,e=n.matchAll(t);return Array.from(e).map(i=>{const o=i.groups.p.split(",");return o.length>=8?{text:i.groups.text.trim().replaceAll("&quot;",'"').replaceAll("&apos;","'").replaceAll("&lt;","<").replaceAll("&gt;",">").replaceAll("&amp;","&"),time:Number(o[0]),mode:Ie(Number(o[1])),fontSize:Number(o[2]),color:`#${Number(o[3]).toString(16)}`,timestamp:Number(o[4]),pool:Number(o[5]),userID:o[6],rowID:Number(o[7])}:null}).filter(Boolean)}function vo({data:n}){const{xml:t,id:e}=n;if(!e||!t)return;const i=jt(t);globalThis.postMessage({danmus:i,id:e})}function yo(){const n=`
        ${Ie.toString()}
        ${jt.toString()}
        onmessage = ${vo.toString()}
    `,t=new Blob([n],{type:"application/javascript"});return new Worker(URL.createObjectURL(t))}function bo(n){return new Promise(async t=>{const i=await(await fetch(n)).text();try{const o=yo();o.onmessage=r=>{const{danmus:a,id:s}=r.data;!s||!a||(t(a),o.terminate())},o.postMessage({xml:i,id:Date.now()})}catch(o){console.error("Error parsing Bilibili Danmu:",o);const r=jt(i);t(r)}})}const Ae=`/*!
 * artplayer-plugin-danmuku.js v5.3.0
 * Github: https://github.com/zhw2590582/ArtPlayer
 * (c) 2017-2026 Harvey Zhao
 * Released under the MIT License.
 */
function getDanmuTop({ target, visibles, clientWidth, clientHeight, marginBottom, marginTop, antiOverlap }) {
  const maxTop = clientHeight - marginBottom;
  const danmus = visibles.filter((item) => item.mode === target.mode && item.top <= maxTop).sort((prev, next) => prev.top - next.top);
  if (danmus.length === 0) {
    if (target.mode === 2) {
      return maxTop - target.height;
    } else {
      return marginTop;
    }
  }
  danmus.unshift({
    type: "top",
    top: 0,
    left: 0,
    right: 0,
    height: marginTop,
    width: clientWidth,
    speed: 0,
    distance: clientWidth
  });
  danmus.push({
    type: "bottom",
    top: maxTop,
    left: 0,
    right: 0,
    height: marginBottom,
    width: clientWidth,
    speed: 0,
    distance: clientWidth
  });
  if (target.mode === 2) {
    for (let index = danmus.length - 2; index >= 0; index -= 1) {
      const item = danmus[index];
      const prev = danmus[index + 1];
      const itemBottom = item.top + item.height;
      const diff = prev.top - itemBottom;
      if (diff >= target.height) {
        return prev.top - target.height;
      }
    }
  } else {
    for (let index = 1; index < danmus.length; index += 1) {
      const item = danmus[index];
      const prev = danmus[index - 1];
      const prevBottom = prev.top + prev.height;
      const diff = item.top - prevBottom;
      if (diff >= target.height) {
        return prevBottom;
      }
    }
  }
  const topMap = [];
  for (let index = 1; index < danmus.length - 1; index += 1) {
    const item = danmus[index];
    if (topMap.length) {
      const last = topMap[topMap.length - 1];
      if (last[0].top === item.top) {
        last.push(item);
      } else {
        topMap.push([item]);
      }
    } else {
      topMap.push([item]);
    }
  }
  if (antiOverlap) {
    switch (target.mode) {
      case 0: {
        const result = topMap.find((list) => {
          return list.every((danmu) => {
            if (clientWidth < danmu.distance)
              return false;
            if (target.speed < danmu.speed)
              return true;
            const overlapTime = danmu.right / (target.speed - danmu.speed);
            if (overlapTime > danmu.time)
              return true;
            return false;
          });
        });
        return result && result[0] ? result[0].top : void 0;
      }
      // 静止弹幕没有重叠问题
      case 1:
      case 2:
        return void 0;
    }
  } else {
    switch (target.mode) {
      case 0:
        topMap.sort((prev, next) => {
          const nextMinRight = Math.min(...next.map((item) => item.right));
          const prevMinRight = Math.min(...prev.map((item) => item.right));
          return nextMinRight * next.length - prevMinRight * prev.length;
        });
        break;
      case 1:
      case 2:
        topMap.sort((prev, next) => {
          const nextMaxWidth = Math.max(...next.map((item) => item.width));
          const prevMaxWidth = Math.max(...prev.map((item) => item.width));
          return prevMaxWidth * prev.length - nextMaxWidth * next.length;
        });
        break;
    }
    return topMap[0][0].top;
  }
}
onmessage = (event) => {
  const { data } = event;
  if (!data.id || !data.type)
    return;
  const fns = { getDanmuTop };
  const fn = fns[data.type];
  const result = fn(data);
  globalThis.postMessage({
    result,
    id: data.id
  });
};
`,ne=typeof self<"u"&&self.Blob&&new Blob(["URL.revokeObjectURL(import.meta.url);",Ae],{type:"text/javascript;charset=utf-8"});function xo(n){let t;try{if(t=ne&&(self.URL||self.webkitURL).createObjectURL(ne),!t)throw"";const e=new Worker(t,{type:"module",name:n?.name});return e.addEventListener("error",()=>{(self.URL||self.webkitURL).revokeObjectURL(t)}),e}catch{return new Worker("data:text/javascript;charset=utf-8,"+encodeURIComponent(Ae),{type:"module",name:n?.name})}}class K{constructor(t,e){const{constructor:i,template:o}=t;this.utils=i.utils,this.validator=i.validator,this.$danmuku=o.$danmuku,this.$player=o.$player,this.art=t,this.queue=[],this.$refs=[],this.isStop=!1,this.isHide=!1,this.timer=null,this.index=0,this.option=K.option,this.states={wait:[],ready:[],emit:[],stop:[]},this.config(e,!0),this.worker=new xo,this.start=this.start.bind(this),this.stop=this.stop.bind(this),this.reset=this.reset.bind(this),this.resize=this.resize.bind(this),this.destroy=this.destroy.bind(this),t.on("video:play",this.start),t.on("video:playing",this.start),t.on("video:pause",this.stop),t.on("video:waiting",this.stop),t.on("destroy",this.destroy),t.on("resize",this.resize),this.load()}static get option(){return{danmuku:[],speed:5,margin:[10,"25%"],opacity:1,color:"#FFFFFF",mode:0,modes:[0,1,2],fontSize:25,antiOverlap:!0,synchronousPlayback:!1,mount:void 0,heatmap:!1,width:512,points:[],filter:()=>!0,beforeEmit:()=>!0,beforeVisible:()=>!0,visible:!0,emitter:!0,maxLength:200,lockTime:5,theme:"dark",OPACITY:{},FONT_SIZE:{},MARGIN:{},SPEED:{},COLOR:[]}}static get scheme(){return{danmuku:"array|function|string",speed:"number",margin:"array",opacity:"number",color:"string",mode:"number",modes:"array",fontSize:"number|string",antiOverlap:"boolean",synchronousPlayback:"boolean",mount:"?htmldivelement|string",heatmap:"object|boolean",width:"number",points:"array",filter:"function",beforeEmit:"function",beforeVisible:"function",visible:"boolean",emitter:"boolean",maxLength:"number",lockTime:"number",theme:"string",OPACITY:"object",FONT_SIZE:"object",MARGIN:"object",SPEED:"object",COLOR:"array"}}static get cssText(){return`
            user-select: none;
            position: absolute;
            white-space: pre;
            pointer-events: none;
            perspective: 500px;
            display: inline-block;
            will-change: transform;
            font-weight: normal;
            line-height: 1.125;
            visibility: hidden;
            font-family: SimHei, "Microsoft JhengHei", Arial, Helvetica, sans-serif;
            text-shadow: rgb(0, 0, 0) 1px 0px 1px, rgb(0, 0, 0) 0px 1px 1px, rgb(0, 0, 0) 0px -1px 1px, rgb(0, 0, 0) -1px 0px 1px;
        `}get isRotate(){return this.art.plugins?.autoOrientation?.state}get marginTop(){const{clamp:t}=this.utils,e=this.option.margin[0],{clientHeight:i}=this.$player;if(typeof e=="number")return t(e,0,i);if(typeof e=="string"&&e.endsWith("%")){const o=Number.parseFloat(e)/100;return t(i*o,0,i)}return K.option.margin[0]}get marginBottom(){const{clamp:t}=this.utils,e=this.option.margin[1],{clientHeight:i}=this.$player;if(typeof e=="number")return t(e,0,i);if(typeof e=="string"&&e.endsWith("%")){const o=Number.parseFloat(e)/100;return t(i*o,0,i)}return K.option.margin[1]}get fontSize(){const{clamp:t}=this.utils,{clientHeight:e}=this.$player,i=this.option.fontSize;if(typeof i=="number")return Math.round(t(i,12,e));if(typeof i=="string"&&i.endsWith("%")){const o=Number.parseFloat(i)/100;return Math.round(t(e*o,12,e))}return K.option.fontSize}get $ref(){const t=this.$refs.pop()||document.createElement("div");return t.style.cssText=K.cssText,t.dataset.mode="",t.dataset.id="",t.className="",t}get readys(){const{currentTime:t}=this.art,e=[];return this.filter("ready",i=>e.push(i)),this.filter("wait",i=>{t+.1>=i.time&&i.time>=t-.1&&e.push(i)}),e}get visibles(){const t=[],{clientWidth:e}=this.$player,i=this.getLeft(this.$player);return this.filter("emit",o=>{const r=o.$ref.offsetTop,a=this.getLeft(o.$ref)-i,s=o.$ref.clientHeight,c=o.$ref.clientWidth,l=a+c,d=e-l,p=l/o.$restTime,u={};u.top=r,u.left=a,u.height=s,u.width=c,u.right=d,u.speed=p,u.distance=l,u.time=o.$restTime,u.mode=o.mode,t.push(u)}),t}get speed(){return this.option.synchronousPlayback&&this.art.playbackRate?this.option.speed/Number(this.art.playbackRate):this.option.speed}async load(t){const{errorHandle:e}=this.utils;let i=[];const o=t||this.option.danmuku;try{typeof o=="function"?i=await o():o instanceof Promise?i=await o:typeof o=="string"?i=await bo(o):Array.isArray(o)&&(i=o),e(Array.isArray(i),"Danmuku need return an array as result"),t===void 0&&(this.reset(),this.queue=[],this.states={wait:[],ready:[],emit:[],stop:[]},this.$refs=[],this.$danmuku.textContent="");for(let r=0;r<i.length;r++){const a=i[r];await this.emit(a)}this.art.emit("artplayerPluginDanmuku:loaded",this.queue)}catch(r){throw this.art.emit("artplayerPluginDanmuku:error",r),r}return this}async emit(t){const{clamp:e}=this.utils;if(this.validator(t,{id:"?string",text:"string",mode:"?number",color:"?string",time:"?number",border:"?boolean",style:"?object"}),!t.text.trim())return this;if(t.time?t.time=e(t.time,0,1/0):t.time=this.art.currentTime+.5,t.mode===void 0&&(t.mode=this.option.mode),t.style===void 0&&(t.style={}),t.color===void 0&&(t.color=this.option.color),![0,1,2].includes(t.mode))return this;if(!this.option.filter(t))return this;const i={...t,$state:"wait",$index:this.index++,$ref:null,$restTime:0,$lastStartTime:0};return this.setState(i,"wait"),this.queue.push(i),this}config(t,e=!1){const{clamp:i}=this.utils,{$controlsCenter:o}=this.art.template;return!Object.keys(t).some(a=>JSON.stringify(this.option[a])!==JSON.stringify(t[a]))&&!e?this:(this.option=Object.assign({},K.option,this.option,t),this.validator(this.option,K.scheme),this.option.mode=i(this.option.mode,0,2),this.option.speed=i(this.option.speed,1,10),this.option.opacity=i(this.option.opacity,0,1),this.option.lockTime=i(this.option.lockTime,1,60),this.option.maxLength=i(this.option.maxLength,1,1e3),this.option.mount=this.option.mount||o,t.fontSize&&this.reset(),this.option.visible?this.show():this.hide(),this.art.emit("artplayerPluginDanmuku:config",this.option),this)}getLeft(t){const e=t.getBoundingClientRect();return this.isRotate?e.top:e.left}postMessage(t={}){return new Promise(e=>{t.id=Date.now(),this.worker.postMessage(t),this.worker.onmessage=i=>{const{data:o}=i;o.id===t.id&&e(o)}})}filter(t,e){const i=this.states[t]||[];for(let o=0;o<i.length;o++)e(i[o]);return i}setState(t,e){this.states[t.$state]=this.states[t.$state].filter(i=>i!==t),t.$state=e,t.$ref&&(t.$ref.dataset.state=e),this.states[e].push(t)}makeWait(t){this.setState(t,"wait"),t.$ref&&(t.$ref.style.cssText=K.cssText,t.$ref.style.visibility="hidden",t.$ref.style.marginLeft="0px",t.$ref.style.transform="translateX(0px)",t.$ref.style.transition="transform 0s linear 0s",this.$refs.push(t.$ref),t.$ref=null)}update(){const{setStyles:t}=this.utils;return this.timer=window.requestAnimationFrame(async()=>{if(this.art.playing&&!this.isHide){this.filter("emit",i=>{const o=(Date.now()-i.$lastStartTime)/1e3;i.$restTime-=o,i.$lastStartTime=Date.now(),i.$restTime<=0&&this.makeWait(i)});const e=this.readys;for(let i=0;i<e.length;i++){const o=e[i];if(await this.option.beforeVisible(o)){const{clientWidth:a,clientHeight:s}=this.$player;o.$ref=this.$ref,o.$ref.textContent=o.text,this.$danmuku.appendChild(o.$ref),o.$ref.style.opacity=this.option.opacity,o.$ref.style.fontSize=`${this.fontSize}px`,o.$ref.style.color=o.color,o.$ref.style.border=o.border?`1px solid ${o.color}`:null,o.$ref.style.backgroundColor=o.border?"rgb(0 0 0 / 50%)":null,t(o.$ref,o.style),o.$lastStartTime=Date.now(),o.$restTime=this.speed;const c=a+o.$ref.clientWidth,{result:l}=await this.postMessage({type:"getDanmuTop",target:{mode:o.mode,height:o.$ref.clientHeight,speed:c/o.$restTime},visibles:this.visibles,antiOverlap:this.option.antiOverlap,clientWidth:a,clientHeight:s,marginBottom:this.marginBottom,marginTop:this.marginTop});if(o.$ref)if(!this.isStop&&l!==void 0){switch(this.setState(o,"emit"),o.$ref.style.top=`${l}px`,o.$ref.style.visibility="visible",o.$ref.dataset.mode=o.mode,o.$ref.dataset.id=o.id||"",o.mode){case 0:{o.$ref.style.left=`${a}px`,o.$ref.style.marginLeft="0px",o.$ref.style.transform=`translateX(${-c}px)`,o.$ref.style.transition=`transform ${o.$restTime}s linear 0s`;break}case 1:case 2:o.$ref.style.left="50%",o.$ref.style.marginLeft=`-${o.$ref.clientWidth/2}px`;break}this.art.emit("artplayerPluginDanmuku:visible",o)}else this.setState(o,"ready"),this.$refs.push(o.$ref),o.$ref=null}}}this.isStop||this.update()}),this}resize(){const{clientWidth:t}=this.$player;this.filter("stop",e=>{e.mode===0&&(e.$ref.style.left=`${t}px`)}),this.filter("emit",e=>{if(e.$lastStartTime=Date.now(),e.mode===0){const i=t+e.$ref.clientWidth;e.$ref.style.left=`${t}px`,e.$ref.style.transform=`translateX(${-i}px)`,e.$ref.style.transition=`transform ${e.$restTime}s linear 0s`}})}continue(){const{clientWidth:t}=this.$player;return this.filter("stop",e=>{if(this.setState(e,"emit"),e.$lastStartTime=Date.now(),e.mode===0){const i=t+e.$ref.clientWidth;e.$ref.style.transform=`translateX(${-i}px)`,e.$ref.style.transition=`transform ${e.$restTime}s linear 0s`}}),this}suspend(){const{clientWidth:t}=this.$player;return this.filter("emit",e=>{if(this.setState(e,"stop"),e.mode===0){const i=t-(this.getLeft(e.$ref)-this.getLeft(this.$player));e.$ref.style.transform=`translateX(${-i}px)`,e.$ref.style.transition="transform 0s linear 0s"}}),this}stop(){return this.isStop=!0,this.suspend(),window.cancelAnimationFrame(this.timer),this.art.emit("artplayerPluginDanmuku:stop"),this}start(){return this.isStop=!1,this.continue(),this.update(),this.art.emit("artplayerPluginDanmuku:start"),this}reset(){return this.queue.forEach(t=>this.makeWait(t)),this.art.emit("artplayerPluginDanmuku:reset"),this}show(){return this.isHide=!1,this.$danmuku.style.opacity=1,this.option.visible=!0,this.art.emit("artplayerPluginDanmuku:show"),this}hide(){return this.isHide=!0,this.$danmuku.style.opacity=0,this.option.visible=!1,this.art.emit("artplayerPluginDanmuku:hide"),this}destroy(){this.stop(),this.worker.terminate(),this.art.off("video:play",this.start),this.art.off("video:playing",this.start),this.art.off("video:pause",this.stop),this.art.off("video:waiting",this.stop),this.art.off("resize",this.reset),this.art.off("destroy",this.destroy),this.art.emit("artplayerPluginDanmuku:destroy")}}const Pt={map(n,t,e,i,o){return(n-t)*(o-i)/(e-t)+i},range(n,t,e){const i=Math.round(n/e)*e;return Array.from({length:Math.floor((t-n)/e)},(o,r)=>r*e+i)}};function wo(n,t){const e=t[0]-n[0],i=t[1]-n[1];return{length:Math.sqrt(e**2+i**2),angle:Math.atan2(i,e)}}function ko(n,t,e){const{query:i}=n.constructor.utils;n.controls.add({name:"heatmap",position:"top",html:"",style:{position:"absolute",top:"-100px",left:"0px",right:"0px",height:"100px",width:"100%",pointerEvents:"none"},mounted(o){let r=null,a=null;function s(c=[]){if(r=null,a=null,o.innerHTML="",!n.duration||n.option.isLive)return;const l={w:o.offsetWidth,h:o.offsetHeight},d={xMin:0,xMax:l.w,yMin:0,yMax:128,scale:.25,opacity:.2,minHeight:Math.floor(l.h*.05),sampling:Math.floor(l.w/100),smoothing:.2,flattening:.2};typeof e=="object"&&Object.assign(d,e);let p=[];if(Array.isArray(c)&&c.length)p=[...c];else{const E=n.duration/l.w;for(let L=0;L<=l.w;L+=d.sampling){const I=t.queue.filter(({time:V})=>V>L*E&&V<=(L+d.sampling)*E).length;p.push([L,I])}}if(p.length===0)return;const u=p[p.length-1],m=u[0],v=u[1];m!==l.w&&p.push([l.w,v]);const f=p.map(E=>E[1]),b=Math.min(...f),C=Math.max(...f),T=(b+C)/2;for(let E=0;E<p.length;E++){const L=p[E],I=L[1];L[1]=I*(I>T?1+d.scale:1-d.scale)+d.minHeight}const y=(E,L,I,V)=>{const lt=wo(L||E,I||E),Yt=Pt.map(Math.cos(lt.angle)*d.flattening,0,1,1,0),qt=lt.angle*Yt+(V?Math.PI:0),Xt=lt.length*d.smoothing,De=E[0]+Math.cos(qt)*Xt,Oe=E[1]+Math.sin(qt)*Xt;return[De,Oe]},g=(E,L,I)=>{const V=y(I[L-1],I[L-2],E),N=y(E,I[L-1],I[L+1],!0),st=L===I.length-1?" z":"";return`C ${V[0]},${V[1]} ${N[0]},${N[1]} ${E[0]},${E[1]}${st}`},M=p.map(E=>{const L=Pt.map(E[0],d.xMin,d.xMax,0,l.w),I=Pt.map(E[1],d.yMin,d.yMax,l.h,0);return[L,I]}).reduce((E,L,I,V)=>I===0?`M ${V[V.length-1][0]},${l.h} L ${L[0]},${l.h} L ${L[0]},${L[1]}`:`${E} ${g(L,I,V)}`,"");o.innerHTML=`
                    <svg viewBox="0 0 ${l.w} ${l.h}">
                        <defs>
                            <linearGradient id="heatmap-solids" x1="0%" y1="0%" x2="100%" y2="0%">
                                <stop offset="0%" style="stop-color:var(--art-theme);stop-opacity:${d.opacity}" />
                                <stop offset="0%" style="stop-color:var(--art-theme);stop-opacity:${d.opacity}" id="heatmap-start" />
                                <stop offset="0%" style="stop-color:var(--art-progress-color);stop-opacity:1" id="heatmap-stop" />
                                <stop offset="100%" style="stop-color:var(--art-progress-color);stop-opacity:1" />
                            </linearGradient>
                        </defs>
                        <path fill="url(#heatmap-solids)" d="${M}"></path>
                    </svg>
                `,r=i("#heatmap-start",o),a=i("#heatmap-stop",o),r.setAttribute("offset",`${n.played*100}%`),a.setAttribute("offset",`${n.played*100}%`)}n.on("video:timeupdate",()=>{r&&a&&(r.setAttribute("offset",`${n.played*100}%`),a.setAttribute("offset",`${n.played*100}%`))}),n.on("setBar",(c,l)=>{r&&a&&c==="played"&&(r.setAttribute("offset",`${l*100}%`),a.setAttribute("offset",`${l*100}%`))}),n.on("ready",()=>s()),n.on("resize",()=>s()),n.on("artplayerPluginDanmuku:loaded",()=>s()),n.on("artplayerPluginDanmuku:points",c=>s(c))}})}const _t='<svg  class="apd-icon apd-check-off" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" data-pointer="none" viewBox="0 0 32 32" width="32"  height="32" ><path d="M8 6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2H8zm0-2h16c2.21 0 4 1.79 4 4v16c0 2.21-1.79 4-4 4H8c-2.21 0-4-1.79-4-4V8c0-2.21 1.79-4 4-4z" fill="#FFFFFF"></path></svg>',It='<svg class="apd-icon apd-check-on" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" data-pointer="none" viewBox="0 0 32 32" width="32"  height="32" ><path d="m13 18.25-1.8-1.8c-.6-.6-1.65-.6-2.25 0s-.6 1.5 0 2.25l2.85 2.85c.318.318.762.468 1.2.448.438.02.882-.13 1.2-.448l8.85-8.85c.6-.6.6-1.65 0-2.25s-1.65-.6-2.25 0l-7.8 7.8zM8 4h16c2.21 0 4 1.79 4 4v16c0 2.21-1.79 4-4 4H8c-2.21 0-4-1.79-4-4V8c0-2.21 1.79-4 4-4z" fill="#00AEEC"></path></svg>',ie='<svg class="apd-icon apd-config-icon" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" data-pointer="none" viewBox="0 0 24 24" width="24"  height="24" ><path fill-rule="evenodd" d="m15.645 4.881 1.06-1.473a.998.998 0 1 0-1.622-1.166L13.22 4.835a110.67 110.67 0 0 0-1.1-.007h-.131c-.47 0-.975.004-1.515.012L8.783 2.3A.998.998 0 0 0 7.12 3.408l.988 1.484c-.688.019-1.418.042-2.188.069a4.013 4.013 0 0 0-3.83 3.44c-.165 1.15-.245 2.545-.245 4.185 0 1.965.115 3.67.35 5.116a4.012 4.012 0 0 0 3.763 3.363c1.903.094 3.317.141 5.513.141a.988.988 0 0 0 0-1.975 97.58 97.58 0 0 1-5.416-.139 2.037 2.037 0 0 1-1.91-1.708c-.216-1.324-.325-2.924-.325-4.798 0-1.563.076-2.864.225-3.904.14-.977.96-1.713 1.945-1.747 2.444-.087 4.465-.13 6.063-.131 1.598 0 3.62.044 6.064.13.96.034 1.71.81 1.855 1.814.075.524.113 1.962.141 3.065v.002c.005.183.01.07.014-.038.004-.096.008-.189.011-.081a.987.987 0 1 0 1.974-.069c-.004-.105-.007-.009-.011.09-.002.056-.004.112-.007.135l-.002.01a.574.574 0 0 1-.005-.091v-.027c-.03-1.118-.073-2.663-.16-3.276-.273-1.906-1.783-3.438-3.74-3.507-.905-.032-1.752-.058-2.543-.079Zm-3.113 4.703h-1.307v4.643h2.2v.04l.651-1.234c.113-.215.281-.389.482-.509v-.11h.235c.137-.049.283-.074.433-.074h1.553V9.584h-1.264a8.5 8.5 0 0 0 .741-1.405l-1.078-.381c-.24.631-.501 1.23-.806 1.786h-1.503l.686-.305c-.228-.501-.5-.959-.806-1.394l-1.034.348c.294.392.566.839.817 1.35Zm-1.7 5.502h2.16l-.564 1.068h-1.595v-1.068Zm-2.498-1.863.152-1.561h1.96V8.289H7.277v.969h2.048v1.435h-1.84l-.306 3.51h2.254c0 1.155-.043 1.906-.12 2.255-.076.348-.38.523-.925.523-.305 0-.61-.022-.893-.055l.294 1.056.061.005c.282.02.546.039.81.039.991-.065 1.547-.414 1.677-1.046.11-.631.175-1.883.175-3.757H8.334Zm5.09-.8v.85h-1.188v-.85h1.187Zm-1.188-.955h1.187v-.893h-1.187v.893Zm2.322.007v-.893h1.241v.893h-1.241Zm.528 2.757a1.26 1.26 0 0 1 1.087-.627l4.003-.009a1.26 1.26 0 0 1 1.094.63l1.721 2.982c.226.39.225.872-.001 1.263l-1.743 3a1.26 1.26 0 0 1-1.086.628l-4.003.009a1.26 1.26 0 0 1-1.094-.63l-1.722-2.982a1.26 1.26 0 0 1 .002-1.263l1.742-3Zm1.967.858a1.26 1.26 0 0 0-1.08.614l-.903 1.513a1.26 1.26 0 0 0-.002 1.289l.885 1.492c.227.384.64.62 1.086.618l2.192-.005a1.26 1.26 0 0 0 1.08-.615l.904-1.518a1.26 1.26 0 0 0 .001-1.288l-.884-1.489a1.26 1.26 0 0 0-1.086-.616l-2.193.005Zm2.517 2.76a1.4 1.4 0 1 1-2.8 0 1.4 1.4 0 0 1 2.8 0Z" clip-rule="evenodd"></path></svg>',oe='<svg class="apd-icon apd-mode-0-off" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" xml:space="preserve" data-pointer="none" style="enable-background:new 0 0 28 28" viewBox="0 0 28 28" width="28"  height="28" ><path d="M23 15c1.487 0 2.866.464 4 1.255V7a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v14a4 4 0 0 0 4 4h11.674A7 7 0 0 1 23 15zM11 9h6a1 1 0 0 1 0 2h-6a1 1 0 0 1 0-2zm-3 2H6V9h2v2zm4 4h-2v-2h2v2zm2-1a1 1 0 0 1 1-1h1a1 1 0 0 1 0 2h-1a1 1 0 0 1-1-1z" fill="#00AEEC"></path><path d="M26.536 18.464a5 5 0 0 0-7.071 0 5 5 0 0 0 0 7.071 5 5 0 1 0 7.071-7.071zm-5.657 5.657a3 3 0 0 1-.586-3.415l4.001 4.001a3 3 0 0 1-3.415-.586zm4.829-.827-4.001-4.001a3.002 3.002 0 0 1 4.001 4.001z" fill="#00AEEC"></path></svg>',At='<svg class="apd-icon apd-mode-0-on" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" xml:space="preserve" data-pointer="none" style="enable-background:new 0 0 28 28" viewBox="0 0 28 28" width="28"  height="28" ><path d="M23 3H5a4 4 0 0 0-4 4v14a4 4 0 0 0 4 4h18a4 4 0 0 0 4-4V7a4 4 0 0 0-4-4zM11 9h6a1 1 0 0 1 0 2h-6a1 1 0 0 1 0-2zm-3 2H6V9h2v2zm4 4h-2v-2h2v2zm9 0h-6a1 1 0 0 1 0-2h6a1 1 0 0 1 0 2z" fill="#FFFFFF"></path></svg>',re='<svg class="apd-icon apd-mode-1-off" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" xml:space="preserve" data-pointer="none" style="enable-background:new 0 0 28 28" viewBox="0 0 28 28" width="28"  height="28" ><path d="M23 15c1.487 0 2.866.464 4 1.255V7a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v14a4 4 0 0 0 4 4h11.674A7 7 0 0 1 23 15zm-4-8h2v2h-2V7zM9 9H7V7h2v2zm4 0h-2V7h2v2zm2-2h2v2h-2V7z" fill="#00AEEC"></path><path d="M26.536 18.464a5 5 0 0 0-7.071 0 5 5 0 0 0 0 7.071 5 5 0 1 0 7.071-7.071zm-5.657 5.657a3 3 0 0 1-.586-3.415l4.001 4.001a3 3 0 0 1-3.415-.586zm4.829-.827-4.001-4.001a3.002 3.002 0 0 1 4.001 4.001z" fill="#00AEEC"></path></svg>',Rt='<svg class="apd-icon apd-mode-1-on" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" xml:space="preserve" data-pointer="none" style="enable-background:new 0 0 28 28" viewBox="0 0 28 28" width="28"  height="28" ><path d="M23 3H5a4 4 0 0 0-4 4v14a4 4 0 0 0 4 4h18a4 4 0 0 0 4-4V7a4 4 0 0 0-4-4zM9 9H7V7h2v2zm4 0h-2V7h2v2zm4 0h-2V7h2v2zm4 0h-2V7h2v2z" fill="#FFFFFF"></path></svg>',ae='<svg class="apd-icon apd-mode-2-off" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" xml:space="preserve" data-pointer="none" style="enable-background:new 0 0 28 28" viewBox="0 0 28 28" width="28"  height="28" ><path d="M23 15c1.487 0 2.866.464 4 1.255V7a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v14a4 4 0 0 0 4 4h11.674A7 7 0 0 1 23 15zM9 21H7v-2h2v2zm4 0h-2v-2h2v2z" fill="#00AEEC"></path><path d="M26.536 18.464a5 5 0 0 0-7.071 0 5 5 0 0 0 0 7.071 5 5 0 1 0 7.071-7.071zm-5.657 5.657a3 3 0 0 1-.586-3.415l4.001 4.001a3 3 0 0 1-3.415-.586zm4.829-.827-4.001-4.001a3.002 3.002 0 0 1 4.001 4.001z" fill="#00AEEC"></path></svg>',Vt='<svg class="apd-icon apd-mode-2-on" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" xml:space="preserve" data-pointer="none" style="enable-background:new 0 0 28 28" viewBox="0 0 28 28" width="28"  height="28" ><path d="M23 3H5a4 4 0 0 0-4 4v14a4 4 0 0 0 4 4h18a4 4 0 0 0 4-4V7a4 4 0 0 0-4-4zM9 21H7v-2h2v2zm4 0h-2v-2h2v2zm4 0h-2v-2h2v2zm4 0h-2v-2h2v2z" fill="#FFFFFF"></path></svg>',se='<svg class="apd-icon apd-toggle-off" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" data-pointer="none" viewBox="0 0 24 24" width="24"  height="24" ><path fill-rule="evenodd" d="m8.085 4.891-.999-1.499a1.008 1.008 0 0 1 1.679-1.118l1.709 2.566c.54-.008 1.045-.012 1.515-.012h.13c.345 0 .707.003 1.088.007l1.862-2.59a1.008 1.008 0 0 1 1.637 1.177l-1.049 1.46c.788.02 1.631.046 2.53.078 1.958.069 3.468 1.6 3.74 3.507.088.613.13 2.158.16 3.276l.001.027c.01.333.017.63.025.856a.987.987 0 0 1-1.974.069c-.008-.23-.016-.539-.025-.881v-.002c-.028-1.103-.066-2.541-.142-3.065-.143-1.004-.895-1.78-1.854-1.813-2.444-.087-4.466-.13-6.064-.131-1.598 0-3.619.044-6.063.13a2.037 2.037 0 0 0-1.945 1.748c-.15 1.04-.225 2.341-.225 3.904 0 1.874.11 3.474.325 4.798.154.949.95 1.66 1.91 1.708a97.58 97.58 0 0 0 5.416.139.988.988 0 0 1 0 1.975c-2.196 0-3.61-.047-5.513-.141A4.012 4.012 0 0 1 2.197 17.7c-.236-1.446-.351-3.151-.351-5.116 0-1.64.08-3.035.245-4.184A4.013 4.013 0 0 1 5.92 4.96c.761-.027 1.483-.05 2.164-.069Zm4.436 4.707h-1.32v4.63h2.222v.848h-2.618v1.078h2.431a5.01 5.01 0 0 1 3.575-3.115V9.598h-1.276a8.59 8.59 0 0 0 .748-1.42l-1.089-.384a14.232 14.232 0 0 1-.814 1.804h-1.518l.693-.308a8.862 8.862 0 0 0-.814-1.408l-1.045.352c.297.396.572.847.825 1.364Zm-4.18 3.564.154-1.485h1.98V8.289h-3.2v.979h2.067v1.43H7.483l-.308 3.454h2.277c0 1.166-.044 1.925-.12 2.277-.078.352-.386.528-.936.528-.308 0-.616-.022-.902-.055l.297 1.067.062.004c.285.02.551.04.818.04 1.001-.066 1.562-.418 1.694-1.056.11-.638.176-1.903.176-3.795h-2.2Zm7.458.11v-.858h-1.254v.858H15.8Zm-2.376-.858v.858h-1.199v-.858h1.2Zm-1.199-.946h1.2v-.902h-1.2v.902Zm2.321 0v-.902H15.8v.902h-1.254Zm3.517 10.594a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm-.002-1.502a2.5 2.5 0 0 1-2.217-3.657l3.326 3.398a2.49 2.49 0 0 1-1.109.259Zm2.5-2.5c0 .42-.103.815-.286 1.162l-3.328-3.401a2.5 2.5 0 0 1 3.614 2.239Z" clip-rule="evenodd"></path></svg>',le='<svg class="apd-icon apd-toggle-on" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" data-pointer="none" viewBox="0 0 24 24" width="24"  height="24" ><path fill-rule="evenodd" d="M11.989 4.828c-.47 0-.975.004-1.515.012l-1.71-2.566a1.008 1.008 0 0 0-1.678 1.118l.999 1.5c-.681.018-1.403.04-2.164.068a4.013 4.013 0 0 0-3.83 3.44c-.165 1.15-.245 2.545-.245 4.185 0 1.965.115 3.67.35 5.116a4.012 4.012 0 0 0 3.763 3.363l.906.046c1.205.063 1.808.095 3.607.095a.988.988 0 0 0 0-1.975c-1.758 0-2.339-.03-3.501-.092l-.915-.047a2.037 2.037 0 0 1-1.91-1.708c-.216-1.324-.325-2.924-.325-4.798 0-1.563.076-2.864.225-3.904.14-.977.96-1.713 1.945-1.747 2.444-.087 4.465-.13 6.063-.131 1.598 0 3.62.044 6.064.13.96.034 1.71.81 1.855 1.814.075.524.113 1.962.141 3.065v.002c.01.342.017.65.025.88a.987.987 0 1 0 1.974-.068c-.008-.226-.016-.523-.025-.856v-.027c-.03-1.118-.073-2.663-.16-3.276-.273-1.906-1.783-3.438-3.74-3.507-.9-.032-1.743-.058-2.531-.078l1.05-1.46a1.008 1.008 0 0 0-1.638-1.177l-1.862 2.59c-.38-.004-.744-.007-1.088-.007h-.13Zm.521 4.775h-1.32v4.631h2.222v.847h-2.618v1.078h2.618l.003.678c.36.026.714.163 1.01.407h.11v-1.085h2.694v-1.078h-2.695v-.847H16.8v-4.63h-1.276a8.59 8.59 0 0 0 .748-1.42L15.183 7.8a14.232 14.232 0 0 1-.814 1.804h-1.518l.693-.308a8.862 8.862 0 0 0-.814-1.408l-1.045.352c.297.396.572.847.825 1.364Zm-4.18 3.564.154-1.485h1.98V8.294h-3.2v.98H9.33v1.43H7.472l-.308 3.453h2.277c0 1.166-.044 1.925-.12 2.277-.078.352-.386.528-.936.528-.308 0-.616-.022-.902-.055l.297 1.067.062.005c.285.02.551.04.818.04 1.001-.067 1.562-.419 1.694-1.057.11-.638.176-1.903.176-3.795h-2.2Zm7.458.11v-.858h-1.254v.858h1.254Zm-2.376-.858v.858h-1.199v-.858h1.2Zm-1.199-.946h1.2v-.902h-1.2v.902Zm2.321 0v-.902h1.254v.902h-1.254Z" clip-rule="evenodd"></path><path fill="#00AEEC" fill-rule="evenodd" d="M22.846 14.627a1 1 0 0 0-1.412.075l-5.091 5.703-2.216-2.275-.097-.086-.008-.005a1 1 0 0 0-1.322 1.493l2.963 3.041.093.083.007.005c.407.315 1 .27 1.354-.124l5.81-6.505.08-.102.005-.008a1 1 0 0 0-.166-1.295Z" clip-rule="evenodd"></path></svg>',ce='<svg class="apd-icon apd-style-icon" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" xml:space="preserve" data-pointer="none" style="enable-background:new 0 0 22 22" viewBox="0 0 22 22" width="36"  height="24" ><path d="M17 16H5c-.55 0-1 .45-1 1s.45 1 1 1h12c.55 0 1-.45 1-1s-.45-1-1-1zM6.96 15c.39 0 .74-.24.89-.6l.65-1.6h5l.66 1.6c.15.36.5.6.89.6.69 0 1.15-.71.88-1.34l-3.88-8.97C11.87 4.27 11.46 4 11 4s-.87.27-1.05.69l-3.88 8.97c-.27.63.2 1.34.89 1.34zM11 5.98 12.87 11H9.13L11 5.98z"></path></svg>',$o=`.artplayer-plugin-danmuku {
  display: flex;
  position: relative;
  z-index: 99;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  height: 32px;
  width: 100%;
  color: #fff;
  font-weight: 300;
  flex-shrink: 0;
  gap: 10px;
}
.artplayer-plugin-danmuku .apd-icon {
  cursor: pointer;
  opacity: 0.75;
  transition: all 0.2s ease;
  fill: #fff;
}
.artplayer-plugin-danmuku .apd-icon:hover {
  opacity: 1;
}
.artplayer-plugin-danmuku .apd-config {
  display: flex;
  position: relative;
}
.artplayer-plugin-danmuku .apd-config .apd-config-panel {
  position: absolute;
  bottom: 24px;
  left: 0;
  width: 320px;
  padding: 10px;
  opacity: 0;
  pointer-events: none;
}
.artplayer-plugin-danmuku .apd-config .apd-config-panel .apd-config-panel-inner {
  width: 100%;
  border-radius: 3px;
  background-color: rgba(0, 0, 0, 0.85);
  padding: 10px;
}
.artplayer-plugin-danmuku .apd-config:hover .apd-config-panel {
  opacity: 100;
  pointer-events: all;
}
.artplayer-plugin-danmuku .apd-config-mode,
.artplayer-plugin-danmuku .apd-config-slider,
.artplayer-plugin-danmuku .apd-config-other,
.artplayer-plugin-danmuku .apd-style-mode {
  margin-bottom: 15px;
}
.artplayer-plugin-danmuku .apd-modes {
  display: flex;
  align-items: center;
  margin-top: 5px;
  gap: 20px;
}
.artplayer-plugin-danmuku .apd-modes .apd-mode {
  cursor: pointer;
  text-align: center;
}
.artplayer-plugin-danmuku .apd-modes .apd-mode:hover {
  color: #00a1d6;
}
.artplayer-plugin-danmuku .apd-config-slider {
  display: flex;
  align-items: center;
  gap: 12px;
}
.artplayer-plugin-danmuku .apd-config-slider .apd-value {
  width: 32px;
  text-align: right;
}
.artplayer-plugin-danmuku .apd-slider {
  position: relative;
  flex: 1;
  display: flex;
  height: 20px;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}
.artplayer-plugin-danmuku .apd-slider .apd-slider-line {
  position: relative;
  height: 2px;
  width: 100%;
  overflow: hidden;
  border-radius: 3px;
  background-color: rgba(255, 255, 255, 0.25);
}
.artplayer-plugin-danmuku .apd-slider .apd-slider-points {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.artplayer-plugin-danmuku .apd-slider .apd-slider-points .apd-slider-point {
  width: 2px;
  height: 2px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.5);
}
.artplayer-plugin-danmuku .apd-slider .apd-slider-progress {
  width: 0%;
  height: 100%;
  background-color: #00a1d6;
}
.artplayer-plugin-danmuku .apd-slider .apd-slider-dot {
  position: absolute;
  transform: translateX(-6px);
  left: 0%;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background-color: #00a1d6;
}
.artplayer-plugin-danmuku .apd-slider .apd-slider-steps {
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: absolute;
  bottom: -12px;
  width: calc(100% + 32px);
  color: #777;
}
.artplayer-plugin-danmuku .apd-slider .apd-slider-steps .apd-slider-step {
  flex-shrink: 0;
  width: 36px;
  text-align: center;
  scale: 0.95;
}
.artplayer-plugin-danmuku .apd-config-other {
  display: flex;
  align-items: center;
  gap: 20px;
}
.artplayer-plugin-danmuku .apd-config-other .apd-check-off,
.artplayer-plugin-danmuku .apd-config-other .apd-check-on {
  width: 16px;
  height: 16px;
}
.artplayer-plugin-danmuku .apd-config-other .apd-other {
  display: flex;
  align-items: center;
  cursor: pointer;
  gap: 2px;
}
.artplayer-plugin-danmuku .apd-config-other .apd-other:hover {
  color: #00a1d6;
}
.artplayer-plugin-danmuku .apd-emitter {
  display: flex;
  flex: 1;
  align-items: center;
  height: 100%;
  background-color: rgba(255, 255, 255, 0.25);
  border-radius: 5px;
}
.artplayer-plugin-danmuku .apd-style {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}
.artplayer-plugin-danmuku .apd-style .apd-style-panel {
  position: absolute;
  bottom: 24px;
  left: 0;
  width: 200px;
  padding: 10px;
  opacity: 0;
  pointer-events: none;
}
.artplayer-plugin-danmuku .apd-style .apd-style-panel .apd-style-panel-inner {
  width: 100%;
  border-radius: 3px;
  background-color: rgba(0, 0, 0, 0.85);
  padding: 10px;
}
.artplayer-plugin-danmuku .apd-style:hover .apd-style-panel {
  opacity: 100;
  pointer-events: all;
}
.artplayer-plugin-danmuku .apd-colors {
  display: flex;
  flex-wrap: wrap;
  margin-top: 5px;
  gap: 8px;
}
.artplayer-plugin-danmuku .apd-colors .apd-color {
  width: 16px;
  height: 16px;
  border-radius: 2px;
  cursor: pointer;
}
.artplayer-plugin-danmuku .apd-colors .apd-color.apd-active {
  border: 1px solid black;
  box-shadow: 0 0 0 1px #fff;
}
.artplayer-plugin-danmuku .apd-input {
  outline: none;
  height: 100%;
  flex: 1;
  min-width: 0;
  width: auto;
  border: none;
  line-height: 1;
  color: #fff;
  background-color: transparent;
}
.artplayer-plugin-danmuku .apd-input::placeholder {
  color: rgba(255, 255, 255, 0.5);
}
.artplayer-plugin-danmuku .apd-send {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  width: 60px;
  flex-shrink: 0;
  cursor: pointer;
  text-shadow: none;
  border-top-right-radius: 5px;
  border-bottom-right-radius: 5px;
  background-color: #00a1d6;
}
.artplayer-plugin-danmuku .apd-send.apd-lock {
  cursor: not-allowed;
  color: #666;
  background-color: #e7e7e7;
}
.art-controls-center .apd-emitter {
  width: 260px;
  flex: none;
}
.art-fullscreen .artplayer-plugin-danmuku,
.art-fullscreen-web .artplayer-plugin-danmuku {
  height: 38px;
  gap: 16px;
}
.art-fullscreen .artplayer-plugin-danmuku .apd-config-icon,
.art-fullscreen-web .artplayer-plugin-danmuku .apd-config-icon,
.art-fullscreen .artplayer-plugin-danmuku .apd-toggle-off,
.art-fullscreen-web .artplayer-plugin-danmuku .apd-toggle-off,
.art-fullscreen .artplayer-plugin-danmuku .apd-toggle-on,
.art-fullscreen-web .artplayer-plugin-danmuku .apd-toggle-on {
  width: 28px;
  height: 28px;
}
.art-fullscreen .artplayer-plugin-danmuku .apd-emitter,
.art-fullscreen-web .artplayer-plugin-danmuku .apd-emitter {
  width: 400px;
  flex: none;
}
.art-video-player > .artplayer-plugin-danmuku {
  position: absolute;
  left: 0;
  right: 0;
  bottom: -40px;
  padding: 0 10px;
}
.art-video-player:has(> .artplayer-plugin-danmuku) {
  margin-bottom: 40px;
}
[data-danmuku-emitter='false'] .apd-emitter {
  display: none !important;
}
[data-danmuku-emitter='false'] .art-controls-center .artplayer-plugin-danmuku {
  justify-content: flex-end;
  gap: 18px;
}
[data-danmuku-emitter='false'].art-fullscreen .art-controls-center .artplayer-plugin-danmuku,
[data-danmuku-emitter='false'].art-fullscreen-web .art-controls-center .artplayer-plugin-danmuku {
  gap: 24px;
}
[data-danmuku-theme='light'] > .artplayer-plugin-danmuku .apd-icon {
  fill: #333;
}
[data-danmuku-theme='light'] > .artplayer-plugin-danmuku .apd-emitter {
  background-color: #f1f2f3;
}
[data-danmuku-theme='light'] > .artplayer-plugin-danmuku .apd-input {
  color: #000;
}
[data-danmuku-theme='light'] > .artplayer-plugin-danmuku .apd-input::placeholder {
  color: rgba(0, 0, 0, 0.3);
}
[data-danmuku-visible='false'] .apd-toggle-off {
  display: block;
}
[data-danmuku-visible='false'] .apd-toggle-on {
  display: none;
}
[data-danmuku-visible='true'] .apd-toggle-off {
  display: none;
}
[data-danmuku-visible='true'] .apd-toggle-on {
  display: block;
}
[data-danmuku-anti-overlap='false'] .apd-anti-overlap .apd-check-on {
  display: none;
}
[data-danmuku-anti-overlap='false'] .apd-anti-overlap .apd-check-off {
  display: block;
}
[data-danmuku-anti-overlap='true'] .apd-anti-overlap .apd-check-on {
  display: block;
}
[data-danmuku-anti-overlap='true'] .apd-anti-overlap .apd-check-off {
  display: none;
}
[data-danmuku-sync-video='false'] .apd-sync-video .apd-check-on {
  display: none;
}
[data-danmuku-sync-video='false'] .apd-sync-video .apd-check-off {
  display: block;
}
[data-danmuku-sync-video='true'] .apd-sync-video .apd-check-on {
  display: block;
}
[data-danmuku-sync-video='true'] .apd-sync-video .apd-check-off {
  display: none;
}
[data-danmuku-mode0='false'] .apd-config-mode .apd-mode-0-off {
  display: block;
}
[data-danmuku-mode0='false'] .apd-config-mode .apd-mode-0-on {
  display: none;
}
[data-danmuku-mode0='false'] .art-danmuku [data-mode='0'] {
  opacity: 0 !important;
}
[data-danmuku-mode0='true'] .apd-config-mode .apd-mode-0-off {
  display: none;
}
[data-danmuku-mode0='true'] .apd-config-mode .apd-mode-0-on {
  display: block;
}
[data-danmuku-mode='0'] .apd-style-mode [data-mode='0'] {
  color: #00a1d6;
}
[data-danmuku-mode='0'] .apd-style-mode [data-mode='0'] path {
  fill: #00a1d6;
}
[data-danmuku-mode1='false'] .apd-config-mode .apd-mode-1-off {
  display: block;
}
[data-danmuku-mode1='false'] .apd-config-mode .apd-mode-1-on {
  display: none;
}
[data-danmuku-mode1='false'] .art-danmuku [data-mode='1'] {
  opacity: 0 !important;
}
[data-danmuku-mode1='true'] .apd-config-mode .apd-mode-1-off {
  display: none;
}
[data-danmuku-mode1='true'] .apd-config-mode .apd-mode-1-on {
  display: block;
}
[data-danmuku-mode='1'] .apd-style-mode [data-mode='1'] {
  color: #00a1d6;
}
[data-danmuku-mode='1'] .apd-style-mode [data-mode='1'] path {
  fill: #00a1d6;
}
[data-danmuku-mode2='false'] .apd-config-mode .apd-mode-2-off {
  display: block;
}
[data-danmuku-mode2='false'] .apd-config-mode .apd-mode-2-on {
  display: none;
}
[data-danmuku-mode2='false'] .art-danmuku [data-mode='2'] {
  opacity: 0 !important;
}
[data-danmuku-mode2='true'] .apd-config-mode .apd-mode-2-off {
  display: none;
}
[data-danmuku-mode2='true'] .apd-config-mode .apd-mode-2-on {
  display: block;
}
[data-danmuku-mode='2'] .apd-style-mode [data-mode='2'] {
  color: #00a1d6;
}
[data-danmuku-mode='2'] .apd-style-mode [data-mode='2'] path {
  fill: #00a1d6;
}
`;class Re{constructor(t,e){this.art=t,this.danmuku=e,this.utils=t.constructor.utils;const{setStyle:i}=this.utils,{$controlsCenter:o}=t.template;i(o,"display","flex"),this.template={$controlsCenter:o,$mount:o,$danmuku:null,$toggle:null,$config:null,$configPanel:null,$configModes:null,$style:null,$stylePanel:null,$styleModes:null,$colors:null,$opacitySlider:null,$opacityValue:null,$marginSlider:null,$marginValue:null,$fontSizeSlider:null,$fontSizeValue:null,$speedSlider:null,$speedValue:null,$input:null,$send:null},this.slider={opacity:null,margin:null,fontSize:null,speed:null},this.emitting=!1,this.isLock=!1,this.timer=null,this.createTemplate(),this.createSliders(),this.createEvents(),this.mount(this.option.mount),t.on("resize",()=>this.resize()),t.on("fullscreen",r=>this.onFullscreen(r)),t.on("fullscreenWeb",r=>this.onFullscreen(r)),t.proxy(this.template.$config,"mouseenter",()=>{this.onMouseEnter({$control:this.template.$config,$panel:this.template.$configPanel})}),t.proxy(this.template.$style,"mouseenter",()=>{this.onMouseEnter({$control:this.template.$style,$panel:this.template.$stylePanel})})}static get icons(){return{$on:le,$off:se,$config:ie,$style:ce,$mode_0_off:oe,$mode_0_on:At,$mode_1_off:re,$mode_1_on:Rt,$mode_2_off:ae,$mode_2_on:Vt,$check_on:It,$check_off:_t}}get option(){return this.danmuku.option}get outside(){return this.template.$mount!==this.template.$controlsCenter}get TEMPLATE(){const{option:t}=this;return`
            <div class="apd-toggle">
                ${le}${se}
            </div>
            <div class="apd-config">
                ${ie}
                <div class="apd-config-panel">
                    <div class="apd-config-panel-inner">
                        <div class="apd-config-mode">
                            按类型屏蔽
                            <div class="apd-modes">
                                <div data-mode="0" class="apd-mode">
                                    ${oe}${At}
                                    <div>滚动</div>
                                </div>
                                <div data-mode="1" class="apd-mode">
                                    ${re}${Rt}
                                    <div>顶部</div>
                                </div>
                                <div data-mode="2" class="apd-mode">
                                    ${ae}${Vt}
                                    <div>底部</div>
                                </div>
                            </div>
                        </div>
                        <div class="apd-config-other">
                            <div class="apd-other apd-anti-overlap">
                                ${It}${_t}
                                防止弹幕重叠
                            </div>
                            <div class="apd-other apd-sync-video">
                                ${It}${_t}
                                同步视频速度
                            </div>
                        </div>
                        <div class="apd-config-slider apd-config-opacity">
                            不透明度
                            <div class="apd-slider"></div>
                            <div class="apd-value">未知</div>
                        </div>
                        <div class="apd-config-slider apd-config-margin">
                            显示区域
                            <div class="apd-slider"></div>
                            <div class="apd-value">未知</div>
                        </div>
                        <div class="apd-config-slider apd-config-fontSize">
                            弹幕字号
                            <div class="apd-slider"></div>
                            <div class="apd-value">未知</div>
                        </div>
                        <div class="apd-config-slider apd-config-speed">
                            弹幕速度
                            <div class="apd-slider"></div>
                            <div class="apd-value">未知</div>
                        </div>
                    </div>
                </div>
            </div>
            <div class="apd-emitter">
                <div class="apd-style">
                    ${ce}
                    <div class="apd-style-panel">
                        <div class="apd-style-panel-inner">
                            <div class="apd-style-mode">
                                模式
                                <div class="apd-modes">
                                    <div data-mode="0" class="apd-mode">
                                        ${At}
                                        <div>滚动</div>
                                    </div>
                                    <div data-mode="1" class="apd-mode">
                                        ${Rt}
                                        <div>顶部</div>
                                    </div>
                                    <div data-mode="2" class="apd-mode">
                                        ${Vt}
                                        <div>底部</div>
                                    </div>
                                </div>
                            </div>
                            <div class="apd-style-color">
                                颜色
                                <div class="apd-colors">
                                    ${this.COLOR.map(e=>`<div data-color="${e}" class="apd-color" style="background-color: ${e}"></div>`).join("")}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                <input class="apd-input" placeholder="发个友善的弹幕见证当下" autocomplete="off" maxLength="${t.maxLength}" />
                <div class="apd-send">发送</div>
            </div>
        `}get OPACITY(){return{min:0,max:100,steps:[],...this.option.OPACITY}}get FONT_SIZE(){return{min:12,max:120,steps:[],...this.option.FONT_SIZE}}get MARGIN(){return{min:0,max:3,steps:[{name:"1/4",value:[10,"75%"]},{name:"半屏",value:[10,"50%"]},{name:"3/4",value:[10,"25%"]},{name:"满屏",value:[10,10]}],...this.option.MARGIN}}get SPEED(){return{min:0,max:4,steps:[{name:"极慢",value:10},{name:"较慢",value:7.5,hide:!0},{name:"适中",value:5},{name:"较快",value:2.5,hide:!0},{name:"极快",value:1}],...this.option.SPEED}}get COLOR(){return this.option.COLOR.length?this.option.COLOR:["#FE0302","#FF7204","#FFAA02","#FFD302","#FFFF00","#A0EE00","#00CD00","#019899","#4266BE","#89D5FF","#CC0273","#222222","#9B9B9B","#FFFFFF"]}query(t){const{query:e}=this.utils,{$danmuku:i}=this.template;return e(t,i)}append(t,e){const{append:i}=this.utils;[...t.children].includes(e)||i(t,e)}setData(t,e){const{$player:i}=this.art.template,{$mount:o}=this.template;i.dataset[t]=e,this.outside&&(o.dataset[t]=e)}createTemplate(){const{createElement:t,tooltip:e}=this.utils,i=t("div");i.className="artplayer-plugin-danmuku",i.innerHTML=this.TEMPLATE,this.template.$danmuku=i,this.template.$toggle=this.query(".apd-toggle"),this.template.$config=this.query(".apd-config"),this.template.$configPanel=this.query(".apd-config-panel"),this.template.$configModes=this.query(".apd-config-mode .apd-modes"),this.template.$style=this.query(".apd-style"),this.template.$stylePanel=this.query(".apd-style-panel"),this.template.$styleModes=this.query(".apd-style-mode .apd-modes"),this.template.$colors=this.query(".apd-colors"),this.template.$antiOverlap=this.query(".apd-anti-overlap"),this.template.$syncVideo=this.query(".apd-sync-video"),this.template.$opacitySlider=this.query(".apd-config-opacity .apd-slider"),this.template.$opacityValue=this.query(".apd-config-opacity .apd-value"),this.template.$marginSlider=this.query(".apd-config-margin .apd-slider"),this.template.$marginValue=this.query(".apd-config-margin .apd-value"),this.template.$fontSizeSlider=this.query(".apd-config-fontSize .apd-slider"),this.template.$fontSizeValue=this.query(".apd-config-fontSize .apd-value"),this.template.$speedSlider=this.query(".apd-config-speed .apd-slider"),this.template.$speedValue=this.query(".apd-config-speed .apd-value"),this.template.$input=this.query(".apd-input"),this.template.$send=this.query(".apd-send");const{$toggle:o}=this.template;this.art.on("artplayerPluginDanmuku:show",()=>{e(o,"关闭弹幕")}),this.art.on("artplayerPluginDanmuku:hide",()=>{e(o,"打开弹幕")})}createEvents(){const{$toggle:t,$configModes:e,$styleModes:i,$colors:o,$antiOverlap:r,$syncVideo:a,$send:s,$input:c}=this.template;this.art.proxy(t,"click",()=>{this.danmuku.config({visible:!this.option.visible}),this.reset()}),this.art.proxy(e,"click",l=>{const d=l.target.closest(".apd-mode");if(!d)return;const p=Number(d.dataset.mode);this.option.modes.includes(p)?this.danmuku.config({modes:this.option.modes.filter(u=>u!==p)}):this.danmuku.config({modes:[...this.option.modes,p]}),this.reset()}),this.art.proxy(r,"click",()=>{this.danmuku.config({antiOverlap:!this.option.antiOverlap}),this.reset()}),this.art.proxy(a,"click",()=>{this.danmuku.config({synchronousPlayback:!this.option.synchronousPlayback}),this.reset()}),this.art.proxy(i,"click",l=>{const d=l.target.closest(".apd-mode");if(!d)return;const p=Number(d.dataset.mode);this.danmuku.config({mode:p}),this.reset()}),this.art.proxy(o,"click",l=>{const d=l.target.closest(".apd-color");d&&(this.danmuku.config({color:d.dataset.color}),this.reset())}),this.art.proxy(s,"click",()=>this.emit()),this.art.proxy(c,"keypress",l=>{l.key==="Enter"&&(l.preventDefault(),this.emit())})}createSliders(){this.slider.opacity=this.createSlider({...this.OPACITY,container:this.template.$opacitySlider,findIndex:()=>Math.round(this.option.opacity*100),onChange:t=>{const{$opacityValue:e}=this.template;e.textContent=`${t}%`,this.danmuku.config({opacity:t/100})}}),this.slider.margin=this.createSlider({...this.MARGIN,container:this.template.$marginSlider,findIndex:()=>this.MARGIN.steps.findIndex(t=>t.value[0]===this.option.margin[0]&&t.value[1]===this.option.margin[1]),onChange:t=>{const e=this.MARGIN.steps[t];if(!e)return;const{$marginValue:i}=this.template;i.textContent=e.name,this.danmuku.config({margin:e.value})}}),this.slider.fontSize=this.createSlider({...this.FONT_SIZE,container:this.template.$fontSizeSlider,findIndex:()=>this.danmuku.fontSize,onChange:t=>{const{$fontSizeValue:e}=this.template;e.textContent=`${t}px`,t!==this.danmuku.fontSize&&this.danmuku.config({fontSize:t})}}),this.slider.speed=this.createSlider({...this.SPEED,container:this.template.$speedSlider,findIndex:()=>this.SPEED.steps.findIndex(t=>t.value===this.option.speed),onChange:t=>{const e=this.SPEED.steps[t];if(!e)return;const{$speedValue:i}=this.template;i.textContent=e.name,this.danmuku.config({speed:e.value})}})}createSlider({min:t,max:e,container:i,findIndex:o,onChange:r,steps:a=[]}){const{query:s,clamp:c,setStyle:l}=this.utils;l(i,"touch-action","none"),i.innerHTML=`
            <div class="apd-slider-line">
                <div class="apd-slider-points">
                    ${a.map(()=>'<div class="apd-slider-point"></div>').join("")}
                </div>
                <div class="apd-slider-progress"></div>
            </div>
            <div class="apd-slider-dot"></div>
            <div class="apd-slider-steps">
                ${a.map(f=>f.hide?"":`<div class="apd-slider-step">${f.name}</div>`).join("")}
            </div>
        `;const d=s(".apd-slider-dot",i),p=s(".apd-slider-progress",i);let u=!1;function m(f=o()){if(f<t||f>e)return;const b=(f-t)/(e-t);d.style.left=`${b*100}%`,a.length===0&&(p.style.width=d.style.left),r(f)}function v(f){const{top:b,height:C,left:T,width:y}=i.getBoundingClientRect();if(this.art.isRotate){const g=c(f.clientY-b,0,C),x=Math.round(g/C*(e-t)+t);m(x)}else{const g=c(f.clientX-T,0,y),x=Math.round(g/y*(e-t)+t);m(x)}}return this.art.proxy(i,"click",f=>{v.call(this,f)}),this.art.proxy(i,"pointerdown",f=>{u=f.button===0}),this.art.on("document:pointermove",f=>{u&&v.call(this,f)}),this.art.on("document:pointerup",f=>{u&&(u=!1,v.call(this,f))}),{reset:m}}onFullscreen(t){const{$danmuku:e,$controlsCenter:i,$mount:o}=this.template;this.outside?t?this.append(i,e):this.append(o,e):this.append(i,e)}onMouseEnter({$control:t,$panel:e}){const{$player:i}=this.art.template,o=t.getBoundingClientRect(),r=e.getBoundingClientRect(),a=i.getBoundingClientRect(),s=r.width/2-o.width/2,c=a.left-(o.left-s),l=o.right+s-a.right;c>0?e.style.left=`${-s+c}px`:l>0?e.style.left=`${-s-l}px`:e.style.left=`${-s}px`}async emit(){const{$input:t}=this.template,e=t.value.trim();if(!e.length||this.isLock||this.emitting)return;const i={text:e,mode:this.option.mode,color:this.option.color,time:this.art.currentTime};try{this.emitting=!0;const o=await this.option.beforeEmit(i);if(this.emitting=!1,o!==!0)return;i.border=!0,delete i.time,this.danmuku.emit(i),t.value="",this.lock()}catch(o){console.error("Error emitting danmuku:",o),this.emitting=!1}}lock(){const{addClass:t}=this.utils,{$send:e}=this.template;this.isLock=!0;let i=this.option.lockTime;e.textContent=i,t(e,"apd-lock");const o=()=>{this.timer=setTimeout(()=>{i===0?this.unlock():(i-=1,e.textContent=i,o())},1e3)};o()}unlock(){const{removeClass:t}=this.utils,{$send:e}=this.template;clearTimeout(this.timer),this.isLock=!1,e.textContent="发送",t(e,"apd-lock")}resize(){if(this.outside||this.art.fullscreen||this.art.fullscreenWeb)return;const{$player:t,$controlsCenter:e}=this.art.template,{$danmuku:i}=this.template;this.art.width<this.option.width?this.append(t,i):this.append(e,i)}reset(){const{inverseClass:t,tooltip:e}=this.utils,{$toggle:i,$colors:o}=this.template;this.slider.opacity.reset(),this.slider.margin.reset(),this.slider.fontSize.reset(),this.slider.speed.reset(),this.setData("danmukuVisible",this.option.visible),this.setData("danmukuMode",this.option.mode),this.setData("danmukuColor",this.option.color),this.setData("danmukuMode0",this.option.modes.includes(0)),this.setData("danmukuMode1",this.option.modes.includes(1)),this.setData("danmukuMode2",this.option.modes.includes(2)),this.setData("danmukuAntiOverlap",this.option.antiOverlap),this.setData("danmukuSyncVideo",this.option.synchronousPlayback),this.setData("danmukuTheme",this.option.theme),this.setData("danmukuEmitter",this.option.emitter);const r=o.children,a=Array.from(r).find(s=>s.dataset.color===this.option.color.toUpperCase());a&&t(a,"apd-active"),e(i,this.option.visible?"关闭弹幕":"打开弹幕"),this.resize()}mount(t){const{errorHandle:e}=this.utils,i=typeof t=="string"?document.querySelector(t):t;e(i,`Can not find the mount point: ${t}`),this.append(i,this.template.$danmuku),this.template.$mount=i,this.reset()}}if(typeof document<"u"){const n="artplayer-plugin-danmuku";let t=document.getElementById(n);t||(t=document.createElement("style"),t.id=n,document.readyState==="loading"?document.addEventListener("DOMContentLoaded",()=>{document.head.appendChild(t)}):(document.head||document.documentElement).appendChild(t)),t.textContent=$o}function Ve(n){return t=>{const e=new K(t,n),i=new Re(t,e);return e.option.heatmap&&ko(t,e,e.option.heatmap),{name:"artplayerPluginDanmuku",emit:e.emit.bind(e),load:e.load.bind(e),config:e.config.bind(e),hide:e.hide.bind(e),show:e.show.bind(e),reset:e.reset.bind(e),mount:i.mount.bind(i),get option(){return e.option},get isHide(){return e.isHide},get isStop(){return e.isStop}}}}Ve.icons=Re.icons;const To={__name:"ArtPlayerWithDanmaku",props:{videoSrc:{type:String,required:!0},danmakuFilePath:{type:String,default:""},cid:{type:String,default:""},poster:{type:String,default:""},title:{type:String,default:"视频播放"},autoplay:{type:Boolean,default:!1},width:{type:String,default:"100%"},height:{type:String,default:"100%"}},emits:["ready","error"],setup(n,{expose:t,emit:e}){Fe(u=>({v8809a176:n.width,acc880f0:n.height}));const i=n,o=e,r=A(null),a=A(null),s=A([]),c=async()=>{if(!i.danmakuFilePath&&!i.cid)return console.log("没有提供弹幕文件路径或CID，跳过弹幕加载"),[];try{const u=i.danmakuFilePath||"",m=i.cid||"",v=await Ue(m,u);if(!v||!v.data)return console.warn("弹幕文件内容为空"),[];const f=l(v.data);return s.value=f,a.value&&a.value.plugins.artplayerPluginDanmuku&&a.value.plugins.artplayerPluginDanmuku.config({danmuku:f}),f}catch(u){return console.error("加载弹幕文件失败:",u),[]}},l=u=>{if(!u)return[];const m=[],v=u.split(`
`);return console.log("弹幕文件总行数:",v.length),v.forEach((f,b)=>{if(f.startsWith("Dialogue:"))try{const C=f.split(",");if(C.length>=10){const T=C[1].trim(),y=d(T),g=C[3].trim();let M=C.slice(9).join(","),E="#ffffff",L=/(\\c|{\\c)&H([0-9A-Fa-f]{2,6})&/,I=M.match(L);if(I){let N=I[2];for(console.log(`找到颜色代码: ${N}, 行: ${b}, 文本: ${M}`);N.length<6;)N="0"+N;if(N.length===6){const st=N.substring(0,2),lt=N.substring(2,4);E=`#${N.substring(4,6)}${lt}${st}`,console.log(`转换后颜色: ${E}`)}}let V=0;if(M.includes("\\an8")||M.includes("{\\an8}"))V=1,console.log("找到顶部弹幕: ",M);else if(M.includes("\\an2")||M.includes("{\\an2}"))V=2,console.log("找到底部弹幕: ",M);else if(M.includes("\\pos")||M.includes("{\\pos")){const N=M.match(/\\pos\((\d+),\s*(\d+)\)/)||M.match(/{\\pos\((\d+),\s*(\d+)\)}/);N&&(parseInt(N[2])<540?(V=1,console.log("找到顶部定位弹幕: ",M)):(V=2,console.log("找到底部定位弹幕: ",M)))}else M.includes("\\move")||M.includes("{\\move")?console.log("找到滚动弹幕: ",M):g.toLowerCase().includes("top")?V=1:g.toLowerCase().includes("bottom")&&(V=2);M=M.replace(/{[^}]*}/g,""),M=M.replace(/\\[a-zA-Z0-9]+(&H[0-9A-Fa-f]+&)?/g,""),M=M.replace(/\\[a-zA-Z0-9]+\([^)]*\)/g,""),M=M.trim(),m.push({text:M,time:y,color:E,border:!1,mode:V})}}catch(C){console.warn("解析弹幕行失败:",f,C)}}),console.log(`成功解析${m.length}条弹幕`),m},d=u=>{const m=u.split(":");if(m.length===3){const v=parseInt(m[0]),f=parseInt(m[1]),b=parseFloat(m[2]);return v*3600+f*60+b}return 0},p=async()=>{if(!r.value)return;a.value&&(a.value.destroy(),a.value=null),await c();const u={container:r.value,url:i.videoSrc,title:i.title,poster:i.poster,volume:.7,autoplay:i.autoplay,autoSize:!1,autoMini:!0,loop:!1,flip:!0,playbackRate:!0,aspectRatio:!0,setting:!0,hotkey:!0,pip:!0,fullscreen:!0,fullscreenWeb:!0,subtitleOffset:!0,miniProgressBar:!0,playsInline:!0,lock:!0,fastForward:!0,autoPlayback:!0,theme:"#fb7299",lang:"zh-cn",moreVideoAttr:{crossOrigin:"anonymous"},icons:{loading:'<svg xmlns="http://www.w3.org/2000/svg" width="50" height="50" viewBox="0 0 24 24"><path fill="#fb7299" d="M12,1A11,11,0,1,0,23,12,11,11,0,0,0,12,1Zm0,19a8,8,0,1,1,8-8A8,8,0,0,1,12,20Z" opacity=".25"/><path fill="#fb7299" d="M12,4a8,8,0,0,1,7.89,6.7A1.53,1.53,0,0,0,21.38,12h0a1.5,1.5,0,0,0,1.48-1.75,11,11,0,0,0-21.72,0A1.5,1.5,0,0,0,2.62,12h0a1.53,1.53,0,0,0,1.49-1.3A8,8,0,0,1,12,4Z"><animateTransform attributeName="transform" dur="0.75s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></path></svg>'},customType:{}};s.value.length>0&&(u.plugins=[Ve({danmuku:s.value,speed:5,opacity:.8,fontSize:25,color:"#ffffff",mode:0,margin:[10,"25%"],antiOverlap:!0,useWorker:!0,synchronousPlayback:!0,filter:()=>!0,lockTime:0,maxLength:100,minWidth:200,maxWidth:400,theme:"dark",disableDanmuku:!1,defaultOff:!1,controls:[{name:"danmuku",position:"right",html:"弹幕",tooltip:"显示/隐藏弹幕",style:{padding:"0 10px",fontSize:"14px",fontWeight:"bold"}}]})]);try{a.value=new z(u),a.value.on("ready",()=>{o("ready",a.value)}),a.value.on("error",m=>{console.error("播放器错误:",m),o("error",m)})}catch(m){console.error("初始化播放器失败:",m),o("error",m)}};return wt(()=>i.videoSrc,()=>{a.value?(a.value.switchUrl(i.videoSrc),(i.danmakuFilePath||i.cid)&&c()):p()},{immediate:!1}),wt([()=>i.danmakuFilePath,()=>i.cid],()=>{a.value&&(i.danmakuFilePath||i.cid)&&c()},{immediate:!1}),Ft(()=>{i.videoSrc&&p()}),de(()=>{a.value&&(a.value.destroy(),a.value=null)}),t({player:a,reload:p,loadDanmaku:c}),(u,m)=>(R(),B("div",{ref_key:"artPlayerContainer",ref:r,class:"art-player-container"},null,512))}},Co=$t(To,[["__scopeId","data-v-2a6a1d86"]]),Mo={key:0,class:"fixed inset-0 z-[9999] flex items-center justify-center"},Eo={class:"relative bg-black rounded-lg shadow-xl w-[90%] max-w-4xl max-h-[90vh] z-10 overflow-hidden"},So={class:"absolute top-0 left-0 right-0 p-4 bg-gradient-to-b from-black/80 to-transparent z-10"},zo={class:"text-white text-lg font-medium truncate"},Lo={key:0,class:"text-green-400 text-xs mt-1 flex items-center"},Po={class:"w-full h-full aspect-video max-h-[80vh] relative"},_o={class:"w-full h-full"},Io={key:0,class:"absolute inset-0 flex items-center justify-center bg-black/90"},Ao={class:"text-center p-6"},Ro={class:"text-white/70 mb-4"},Vo=Object.assign({name:"VideoPlayerDialog"},{__name:"VideoPlayerDialog",props:{show:{type:Boolean,default:!1},videoPath:{type:String,default:""}},emits:["update:show"],setup(n,{emit:t}){const e=n,i=t,o=A(null),r=A(!1),a=A(!1),s=A(""),c=A(""),l=A(""),d=A(!1),p=A(""),u=T=>T&&T.split("\\").pop().split("/").pop()||"未知文件",m=T=>{if(!T)return"";const y=u(T),g=y.match(/_(\d+)\.mp4$/)||y.match(/_(\d+)\.flv$/)||y.match(/_(\d+)\.m4a$/);if(g&&g[1])return g[1];const x=T.match(/(\d{8,})/);return x?x[1]:""},v=T=>T?`${T.replace(/\.(mp4|flv|m4a)$/i,"")}.ass`:"",f=()=>{o.value&&o.value.player&&o.value.player.destroy(),s.value="",c.value="",l.value="",a.value=!1},b=()=>{f(),d.value=!1,p.value="",r.value=!1,i("update:show",!1)},C=()=>{e.videoPath&&(f(),l.value=m(e.videoPath),c.value=v(e.videoPath),s.value=Ge(e.videoPath),Be(()=>{a.value=!0}))};return wt(()=>e.show,T=>{T?C():f()}),wt(()=>e.videoPath,T=>{e.show&&T&&C()}),Ft(()=>{e.show&&e.videoPath&&C()}),de(()=>{f()}),(T,y)=>(R(),Dt(ue,{to:"body"},[n.show?(R(),B("div",Mo,[w("div",{class:"fixed inset-0 bg-black/80 backdrop-blur-sm",onClick:b}),w("div",Eo,[w("button",{onClick:b,class:"absolute right-4 top-4 text-white/70 hover:text-white z-20 bg-black/40 p-2 rounded-full"},[...y[0]||(y[0]=[w("svg",{class:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor"},[w("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M6 18L18 6M6 6l12 12"})],-1)])]),w("div",So,[w("h3",zo,Y(u(n.videoPath)),1),c.value?(R(),B("div",Lo,[...y[1]||(y[1]=[w("svg",{class:"w-3.5 h-3.5 mr-1",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor"},[w("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M5 13l4 4L19 7"})],-1),w("span",null,"已加载弹幕",-1)])])):J("",!0)]),w("div",Po,[pe(w("div",_o,[a.value?(R(),Dt(Co,{key:0,ref_key:"artPlayerRef",ref:o,videoSrc:s.value,cid:l.value,danmakuFilePath:c.value,title:u(n.videoPath),autoplay:!0,width:"100%",height:"100%"},null,8,["videoSrc","cid","danmakuFilePath","title"])):J("",!0)],512),[[He,a.value]])]),d.value?(R(),B("div",Io,[w("div",Ao,[y[2]||(y[2]=w("svg",{class:"w-16 h-16 text-red-500 mx-auto mb-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor"},[w("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"})],-1)),y[3]||(y[3]=w("h3",{class:"text-xl font-medium text-white mb-2"},"视频播放失败",-1)),w("p",Ro,Y(p.value),1),w("button",{onClick:b,class:"px-4 py-2 bg-[#fb7299] text-white rounded-md hover:bg-[#fb7299]/90 transition-colors"}," 关闭 ")])])):J("",!0)])])):J("",!0)]))}}),Do=$t(Vo,[["__scopeId","data-v-7ba2581d"]]),Oo={class:"relative"},Fo={class:"flex h-9 items-center rounded-md border border-gray-300 dark:border-gray-600 bg-transparent focus-within:border-[#fb7299] transition-colors duration-200"},Ho=["value","placeholder"],Bo={__name:"SimpleSearchBar",props:{modelValue:{type:String,default:""},placeholder:{type:String,default:"输入关键词搜索..."}},emits:["update:modelValue","search"],setup(n){return(t,e)=>(R(),B("div",Oo,[w("div",Fo,[e[3]||(e[3]=w("div",{class:"pl-3 text-gray-400 dark:text-gray-500"},[w("svg",{xmlns:"http://www.w3.org/2000/svg",class:"h-3.5 w-3.5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor"},[w("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"})])],-1)),w("input",{value:n.modelValue,onInput:e[0]||(e[0]=i=>t.$emit("update:modelValue",i.target.value)),type:"search",placeholder:n.placeholder,class:"h-full w-full border-none bg-transparent px-2 pr-3 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-0 text-xs leading-none",onKeyup:e[1]||(e[1]=Ne(i=>t.$emit("search"),["enter"])),onSearch:e[2]||(e[2]=i=>t.$emit("search"))},null,40,Ho)])]))}},No=$t(Bo,[["__scopeId","data-v-fb1c7a50"]]),Wo={class:"overflow-y-auto"},jo={class:"mb-6"},Yo={key:0,class:"flex flex-col items-center justify-center py-12"},qo={key:1,class:"flex flex-col items-center justify-center py-12"},Xo={key:2},Uo={class:"text-sm text-gray-500 dark:text-gray-400 mb-4"},Go={class:"grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3"},Zo=["onClick"],Ko=["src","alt"],Jo={class:"absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"},Qo=["onClick"],tr={class:"absolute bottom-1 right-1 bg-black/60 backdrop-blur-sm px-1 py-0.5 rounded text-white text-[10px]"},er={key:0},nr=["onClick"],ir={key:0,class:"absolute left-1 top-1 rounded bg-[#fb7299] px-1 py-0.5 text-[10px] text-white"},or={class:"p-2 flex flex-col space-y-1"},rr=["onClick"],ar={class:"flex items-center space-x-1"},sr=["src","alt","onClick"],lr=["onClick"],cr={class:"flex justify-between items-center text-[10px] text-gray-500 dark:text-gray-400"},dr={class:"flex items-center space-x-1"},pr={class:"mt-8 flex justify-center"},ur={key:0,class:"fixed inset-0 z-50 flex items-center justify-center"},hr={class:"relative bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 w-[500px] z-10 p-6"},fr={class:"font-medium text-gray-800 dark:text-gray-100 mb-2"},mr={class:"mb-3 text-sm text-gray-500"},gr={key:0},vr={class:"mt-2"},yr={class:"px-3 py-2 bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 text-sm break-all"},br={class:"mb-4"},xr={class:"flex items-center space-x-2 cursor-pointer select-none"},wr={key:0,class:"mb-4 p-2 bg-amber-50 dark:bg-amber-900/20 rounded-md border border-amber-200 dark:border-amber-800"},kr={class:"flex justify-end space-x-3 mt-6"},$r=["disabled"],Tr=Object.assign({name:"Downloads"},{__name:"Downloads",setup(n){const t=A(!0),e=A({videos:[],total:0,page:1,limit:20,pages:1}),i=A(""),o=A(1),r=A(!1),a=A(""),s=A(!1),c=A(null),l=A(!0),d=A(!1),p=async()=>{try{t.value=!0;const y=await Ze(i.value,o.value,20);y.data&&y.data.status==="success"?e.value={videos:y.data.videos,total:y.data.total,page:y.data.page,limit:y.data.limit,pages:y.data.pages}:console.error("获取下载视频失败:",y.data?.message||"未知错误")}catch(y){console.error("请求获取下载视频列表出错:",y)}finally{t.value=!1}},u=y=>{o.value=y,p()},m=y=>{r.value=!1,a.value="",setTimeout(()=>{a.value=y,r.value=!0},50)},v=y=>{c.value=y,s.value=!0,l.value=!0},f=y=>{if(y.directory)return y.directory;if(y.files&&y.files.length>0){const g=y.files[0].file_path;if(g){const x=Math.max(g.lastIndexOf("/"),g.lastIndexOf("\\"));if(x!==-1)return g.substring(0,x)}}return null},b=async()=>{try{d.value=!0;const y=f(c.value),g=c.value?.cid||null;if(!g&&!y){Lt({type:"warning",message:"无法获取视频信息，删除失败"}),d.value=!1;return}const x=await Ke(g,l.value,y);if(x.data&&x.data.status==="success")Lt({type:"success",message:x.data.message}),s.value=!1,await p();else throw new Error(x.data?.message||"删除视频失败")}catch(y){Lt({type:"danger",message:y.response?.data?.message||y.message||"删除视频失败"})}finally{d.value=!1}},C=async y=>{const g=`https://www.bilibili.com/video/${y.bvid}`;await Gt(g)},T=async y=>{if(y.author_mid){const g=`https://space.bilibili.com/${y.author_mid}`;await Gt(g)}};return Ft(()=>{p()}),(y,g)=>(R(),B("div",Wo,[w("div",jo,[zt(No,{modelValue:i.value,"onUpdate:modelValue":g[0]||(g[0]=x=>i.value=x),placeholder:"搜索已下载的视频或目录路径...",onSearch:p,class:"w-full"},null,8,["modelValue"])]),w("div",null,[t.value?(R(),B("div",Yo,[...g[5]||(g[5]=[w("svg",{class:"animate-spin h-8 w-8 text-[#fb7299] mb-4",fill:"none",viewBox:"0 0 24 24"},[w("circle",{class:"opacity-25",cx:"12",cy:"12",r:"10",stroke:"currentColor","stroke-width":"4"}),w("path",{class:"opacity-75",fill:"currentColor",d:"M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"})],-1),w("p",{class:"text-gray-500 dark:text-gray-400"},"加载中，请稍候...",-1)])])):!e.value.videos||e.value.videos.length===0?(R(),B("div",qo,[...g[6]||(g[6]=[w("svg",{class:"w-16 h-16 text-gray-400 mb-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor"},[w("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"1.5",d:"M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"})],-1),w("p",{class:"text-xl font-medium text-gray-600 dark:text-gray-300 mb-2"},"暂无下载记录",-1),w("p",{class:"text-gray-500 dark:text-gray-400 mb-4 text-center"},' 你还没有下载任何视频，在浏览历史记录时可以点击"下载"按钮下载视频。 ',-1)])])):(R(),B("div",Xo,[w("p",Uo," 共找到 "+Y(e.value.total)+" 个下载记录，当前第 "+Y(e.value.page)+" 页，共 "+Y(e.value.pages)+" 页 ",1),w("div",Go,[(R(!0),B(We,null,je(e.value.videos,(x,M)=>(R(),B("div",{key:M,class:"bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm rounded-md overflow-hidden border border-gray-200/50 dark:border-gray-700/50 hover:border-[#fb7299] hover:shadow-sm transition-all duration-200 relative group"},[w("div",{class:"relative pb-[56.25%] overflow-hidden cursor-pointer group",onClick:E=>C(x)},[w("img",{src:Ut(Zt)(x.cover)||"https://i0.hdslb.com/bfs/archive/c9e72655b7c9c9c68a30d3275313c501e68427d1.jpg",alt:x.title,class:"absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-300",loading:"lazy",onerror:"this.src='https://i0.hdslb.com/bfs/archive/c9e72655b7c9c9c68a30d3275313c501e68427d1.jpg'"},null,8,Ko),w("div",Jo,[x.files&&x.files.length>0&&!x.files[0].is_audio_only?(R(),B("button",{key:0,onClick:vt(E=>m(x.files[0].file_path),["stop"]),class:"w-8 h-8 rounded-full bg-[#fb7299]/80 text-white flex items-center justify-center backdrop-blur-sm"},[...g[7]||(g[7]=[w("svg",{class:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor"},[w("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"}),w("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})],-1)])],8,Qo)):J("",!0)]),w("div",tr,[x.files&&x.files.length>0?(R(),B("div",er,Y(x.files[0].size_mb.toFixed(1))+" MB ",1)):J("",!0)]),w("div",{class:"absolute right-1.5 top-1.5 z-20 hidden group-hover:flex items-center justify-center w-6 h-6 bg-[#7d7c75]/60 backdrop-blur-sm hover:bg-[#7d7c75]/80 rounded-md cursor-pointer transition-all duration-200",onClick:vt(E=>v(x),["stop"])},[...g[8]||(g[8]=[w("svg",{class:"w-3.5 h-3.5 text-white",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor"},[w("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"})],-1)])],8,nr),x.files&&x.files.length>1?(R(),B("div",ir,Y(x.files.length),1)):J("",!0)],8,Zo),w("div",or,[w("div",{class:"line-clamp-1 text-xs text-gray-900 dark:text-gray-100 font-medium cursor-pointer",onClick:E=>C(x)},Y(x.title),9,rr),w("div",ar,[w("img",{src:Ut(Zt)(x.author_face)||"https://i1.hdslb.com/bfs/face/1b6f746be0d0c8324e01e618c5e85e113a8b38be.jpg",alt:x.author_name,class:"w-3.5 h-3.5 rounded-full object-cover cursor-pointer",loading:"lazy",onerror:"this.src='https://i1.hdslb.com/bfs/face/1b6f746be0d0c8324e01e618c5e85e113a8b38be.jpg'",onClick:vt(E=>T(x),["stop"])},null,8,sr),w("span",{class:"text-[10px] text-gray-600 dark:text-gray-400 truncate hover:text-[#fb7299] cursor-pointer",onClick:vt(E=>T(x),["stop"])},Y(x.author_name||"未知UP主"),9,lr)]),w("div",cr,[w("div",dr,[g[9]||(g[9]=w("svg",{class:"w-2.5 h-2.5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor"},[w("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"})],-1)),w("span",null,Y(x.download_date),1)])])])]))),128))]),w("div",pr,[zt(Xe,{"current-page":o.value,"total-pages":e.value.pages,onPageChange:u},null,8,["current-page","total-pages"])])]))]),zt(Do,{show:r.value,"onUpdate:show":g[1]||(g[1]=x=>r.value=x),"video-path":a.value},null,8,["show","video-path"]),(R(),Dt(ue,{to:"body"},[s.value?(R(),B("div",ur,[w("div",{class:"fixed inset-0 bg-black/50",onClick:g[2]||(g[2]=x=>s.value=!1)}),w("div",hr,[g[13]||(g[13]=w("h3",{class:"text-lg font-medium text-gray-900 dark:text-gray-100 mb-4"},"确认删除视频",-1)),g[14]||(g[14]=w("p",{class:"text-gray-600 dark:text-gray-400 mb-4"}," 确定要删除以下视频吗？此操作不可恢复。 ",-1)),w("p",fr,Y(c.value?.title||"未知视频"),1),w("div",mr,[c.value?.cid?(R(),B("p",gr,"CID: "+Y(c.value?.cid),1)):J("",!0),w("div",vr,[g[10]||(g[10]=w("p",{class:"text-gray-600 dark:text-gray-400 mb-1"},"目录路径:",-1)),w("div",yr,Y(f(c.value)||"无法获取目录路径"),1)])]),w("div",br,[w("label",xr,[pe(w("input",{type:"checkbox","onUpdate:modelValue":g[3]||(g[3]=x=>l.value=x),class:"w-4 h-4 text-[#fb7299] border-gray-300 dark:border-gray-600 rounded focus:ring-[#fb7299]"},null,512),[[Ye,l.value]]),g[11]||(g[11]=w("span",null,"同时删除整个目录（包含所有相关文件）",-1))])]),!c.value?.cid&&f(c.value)?(R(),B("div",wr,[...g[12]||(g[12]=[w("p",{class:"text-sm text-amber-700"},[w("span",{class:"font-medium"},"提示："),qe(" 该视频可能来自收藏夹批量下载，将使用目录路径进行删除。 ")],-1)])])):J("",!0),w("div",kr,[w("button",{onClick:g[4]||(g[4]=x=>s.value=!1),class:"px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"}," 取消 "),w("button",{onClick:b,class:"px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700",disabled:d.value||!c.value?.cid&&!f(c.value)},Y(d.value?"删除中...":"确认删除"),9,$r)])])])):J("",!0)]))]))}}),Ir=$t(Tr,[["__scopeId","data-v-c3bdfd2e"]]);export{Ir as D,No as S};
