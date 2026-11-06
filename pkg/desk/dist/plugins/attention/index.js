// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = Array.isArray, i = Array.prototype.indexOf, a = Array.prototype.includes, o = Array.from, s = Object.defineProperty, c = Object.getOwnPropertyDescriptor, l = Object.getOwnPropertyDescriptors, u = Object.prototype, d = Array.prototype, f = Object.getPrototypeOf, p = Object.isExtensible, m = () => {};
function h(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function g() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, ee = 1 << 25, C = 65536, w = 1 << 19, te = 1 << 20, ne = 1 << 25, re = 1 << 21, ie = 1 << 22, ae = 1 << 23, oe = Symbol("$state"), se = Symbol("component"), ce = Symbol("legacy props"), le = Symbol(""), ue = Symbol("attributes"), de = Symbol("class"), fe = Symbol("style"), pe = Symbol("text"), me = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), he = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function ge() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function _e(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function ve() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var T = !1;
function ye(e) {
	T = e;
}
var E;
function D(t) {
	if (t === null) throw _e(), e;
	return E = t;
}
function be() {
	return D(/* @__PURE__ */ Ut(E));
}
function O(t) {
	if (T) {
		if (/* @__PURE__ */ Ut(E) !== null) throw _e(), e;
		E = t;
	}
}
function xe(e = 1) {
	if (T) {
		for (var t = e, n = E; t--;) n = /* @__PURE__ */ Ut(n);
		E = n;
	}
}
function Se(e = !0) {
	for (var t = 0, n = E;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Ut(n);
		e && n.remove(), n = i;
	}
}
function Ce(t) {
	if (!t || t.nodeType !== 8) throw _e(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function we(e) {
	return e === this.v;
}
function Te(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Ee(e) {
	return !Te(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function De() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Oe(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function ke(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Ae() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function je(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Me() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Ne(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function Pe() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Fe() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Ie() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Le() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var k = null;
function Re(e) {
	k = e;
}
function ze(e, t = !1, n) {
	k = {
		p: k,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: G,
		l: null
	};
}
function Be(e) {
	var t = k, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) rn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, k = t.p, Ve(e);
}
function Ve(e = {}) {
	return s(e, se, { value: !0 }), e;
}
function He() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Ue = [];
function We() {
	var e = Ue;
	Ue = [], h(e);
}
function Ge(e) {
	if (Ue.length === 0 && !pt) {
		var t = Ue;
		queueMicrotask(() => {
			t === Ue && We();
		});
	}
	Ue.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Ke = ~(v | y | _);
function A(e, t) {
	e.f = e.f & Ke | t;
}
function qe(e) {
	e.f & 512 || e.deps === null ? A(e, _) : A(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Je(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), A(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Ye(e) {
	var t = U, n = G;
	W(null), Cn(null);
	try {
		return e();
	} finally {
		W(t), Cn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Xe(e, t, n, r) {
	let i = He() ? et : it;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = G, c = Ze(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				Xt(e, s);
			}
			Qe();
		}
	}
	var d = $e();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ nt(e))).then(u).catch((e) => Xt(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), Qe());
	}) : f();
}
function Ze() {
	var e = G, t = U, n = k, r = j;
	return function(i = !0) {
		Cn(e), W(t), Re(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function Qe(e = !0) {
	Cn(null), W(null), Re(null), e && j?.deactivate();
}
function $e() {
	var e = G, t = e.b, n = j, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function et(e) {
	var n = 2 | v;
	return G !== null && (G.f |= w), {
		ctx: k,
		deps: null,
		effects: null,
		equals: we,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: G,
		ac: null
	};
}
var tt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function nt(e, n, r) {
	let i = G;
	i === null && De();
	var a = void 0, o = kt(t), s = !U, c = /* @__PURE__ */ new Set();
	return on(() => {
		var t = G, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== me && n.reject(e);
			}).finally(Qe);
		} catch (e) {
			n.reject(e), Qe();
		}
		var r = j;
		if (s) {
			if (t.f & 32768) var l = $e();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(tt);
			else for (let e of c.values()) e.reject(tt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== tt && (r.activate(), t ? (o.f |= ae, Nt(o, t)) : (o.f & 8388608 && (o.f ^= ae), Nt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), tn(() => {
		for (let e of c) e.reject(tt);
	}), new Promise((e) => {
		function t(n) {
			function r() {
				n === a ? e(o) : t(a);
			}
			n.then(r, r);
		}
		t(a);
	});
}
/*#__NO_SIDE_EFFECTS__*/
function rt(e) {
	let t = /* @__PURE__ */ et(e);
	return Tn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function it(e) {
	let t = /* @__PURE__ */ et(e);
	return t.equals = Ee, t;
}
function at(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) H(t[n]);
	}
}
function ot(e) {
	var n, r = G, i = e.parent;
	if (!bn && i !== null && e.v !== t && i.f & 24576) return ge(), e.v;
	Cn(i);
	try {
		at(e), n = Pn(e);
	} finally {
		Cn(r);
	}
	return n;
}
function st(e) {
	var t = ot(e);
	!e.equals(t) && (e.wv = jn(), (!j?.is_fork || e.deps === null) && (j === null ? e.v = t : (j.capture(e, t, !0), dt?.capture(e, t, !0)), e.deps === null)) ? A(e, _) : bn || (M === null ? qe(e) : (en() || j?.is_fork) && M.set(e, t));
}
function ct(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Ye(() => {
		t.ac.abort(me), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Ln(t, 0), un(t));
}
function lt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Rn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var ut = null, j = null, dt = null, M = null, ft = null, pt = !1, mt = !1, ht = null, gt = null, _t = 0, vt = 1, yt = class e {
	id = vt++;
	started = !1;
	linked = !0;
	prev = null;
	next = null;
	async_deriveds = /* @__PURE__ */ new Map();
	current = /* @__PURE__ */ new Map();
	previous = /* @__PURE__ */ new Map();
	#e = /* @__PURE__ */ new Set();
	#t = /* @__PURE__ */ new Set();
	#n = 0;
	#r = /* @__PURE__ */ new Map();
	#i = null;
	#a = [];
	#o = [];
	#s = /* @__PURE__ */ new Map();
	#c = /* @__PURE__ */ new Map();
	#l = /* @__PURE__ */ new Set();
	is_fork = !1;
	#u = !1;
	constructor() {
		ut === null ? ut = this : (ut.next = this, this.prev = ut), ut = this;
	}
	#d() {
		if (this.is_fork) return !0;
		for (let e of this.#r.keys()) if (!this.#f(e)) return !0;
		return !1;
	}
	#f(e) {
		for (var t = e; t !== null; t = t.parent) if (this.#c.has(t)) return !0;
		return !1;
	}
	skip_effect(e) {
		this.#c.has(e) || this.#c.set(e, {
			d: [],
			m: []
		}), this.#l.delete(e);
	}
	unskip_effect(e, t = (e) => this.schedule(e)) {
		var n = this.#c.get(e);
		if (n) {
			this.#c.delete(e);
			for (var r of n.d) A(r, v), t(r);
			for (r of n.m) A(r, y), t(r);
		}
		this.#l.add(e);
	}
	#p() {
		var e = [];
		for (let i of this.#a) if (!(i.f & 16384 || !(i.f & 6144))) {
			for (var t = i, n = !1; t.parent !== null;) {
				t = t.parent;
				var r = t.f;
				if (r & 96) {
					if (!(r & 1024)) {
						n = !0;
						break;
					}
					t.f ^= _;
				}
			}
			n || e.push(t);
		}
		return this.#a = [], e;
	}
	#m() {
		this.started = !0;
		for (let [e, t] of this.#s) A(e, t), this.schedule(e);
		this.apply();
		for (var t = ht = [], n = [], r = gt = []; this.#a.length > 0;) {
			_t++ > 1e3 && (this.#b(), bt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Tt(e), this.#d() || this.discard(), t;
			}
		}
		if (j = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (ht = null, gt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) wt(e, t);
			r.length > 0 && j.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), dt = this, St(n), St(t), dt = null, this.#i?.resolve();
			var o = j;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Dt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : Mn(r) && (i & 16 && this.add_dirty_reaction(r, y), Rn(r));
				var o = r.first;
				if (o !== null) {
					r = o;
					continue;
				}
			}
			for (; r !== null;) {
				var s = r.next;
				if (s !== null) {
					r = s;
					break;
				}
				r = r.parent;
			}
		}
	}
	#g() {
		for (var e = this.prev; e !== null; e = e.prev) if (!e.is_fork && this.#_(e)) return e;
		return null;
	}
	#_(e) {
		for (let [t, [, n]] of this.current) if (!n && e.current.has(t)) return !0;
		return !1;
	}
	#v(e) {
		for (let [t, n] of e.current) !this.previous.has(t) && e.previous.has(t) && this.previous.set(t, e.previous.get(t)), this.current.set(t, n);
		for (let [t, n] of e.async_deriveds) {
			let e = this.async_deriveds.get(t);
			e && n.promise.then(e.resolve).catch(e.reject);
		}
		e.async_deriveds.clear(), this.transfer_reactions(e.#s);
		let t = (e) => {
			var n = e.reactions;
			if (n !== null && !(e.f & 2 && !(e.f & 6144))) for (let e of n) {
				var r = e.f;
				if (r & 2) t(e);
				else {
					var i = e;
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), A(i, v), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), j = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Je(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), M?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		j = this;
	}
	deactivate() {
		j = null, M = null;
	}
	flush() {
		try {
			mt = !0, j = this, this.#m();
		} finally {
			_t = 0, ft = null, ht = null, gt = null, mt = !1, j = null, M = null, Dt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(tt);
		this.#b(), this.#i?.resolve();
	}
	register_created_effect(e) {
		this.#o.push(e);
	}
	increment(e, t) {
		if (this.#n += 1, e) {
			let e = this.#r.get(t) ?? 0;
			this.#r.set(t, e + 1);
		}
	}
	decrement(e, t) {
		if (--this.#n, e) {
			let e = this.#r.get(t) ?? 0;
			e === 1 ? this.#r.delete(t) : this.#r.set(t, e - 1);
		}
		this.#u || (this.#u = !0, Ge(() => {
			this.#u = !1, this.linked && this.flush();
		}));
	}
	transfer_reactions(e) {
		for (let [t, n] of e) this.add_dirty_reaction(t, n);
		e.clear();
	}
	add_dirty_reaction(e, t) {
		(t === 2048 || this.#s.get(e) !== 2048) && this.#s.set(e, t);
	}
	oncommit(e) {
		this.#e.add(e);
	}
	ondiscard(e) {
		this.#t.add(e);
	}
	settled() {
		return (this.#i ??= g()).promise;
	}
	static ensure() {
		if (j === null) {
			let t = j = new e();
			!mt && Ge(() => {
				t.started || t.flush();
			});
		}
		return j;
	}
	apply() {
		M = null;
	}
	schedule(e) {
		ft = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? ut = e : t.prev = e, this.linked = !1;
		}
	}
};
function bt() {
	try {
		Me();
	} catch (e) {
		Xt(e, ft);
	}
}
var xt = null;
function St(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Mn(r) && (xt = /* @__PURE__ */ new Set(), Rn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && pn(r), xt?.size > 0)) {
				Dt.clear();
				for (let e of xt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) xt.has(n) && (xt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Rn(n);
					}
				}
				xt.clear();
			}
		}
		xt = null;
	}
}
function Ct(e) {
	j.schedule(e);
}
function wt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), A(e, _);
		for (var n = e.first; n !== null;) wt(n, t), n = n.next;
	}
}
function Tt(e) {
	A(e, _);
	for (var t = e.first; t !== null;) Tt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Et = /* @__PURE__ */ new Set(), Dt = /* @__PURE__ */ new Map(), Ot = !1;
function kt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: we,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function N(e, t) {
	let n = kt(e, t);
	return Tn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function At(e, t = !1, n = !0) {
	let r = kt(e);
	return t || (r.equals = Ee), r;
}
function P(e, t, n = !1) {
	return U !== null && (!Sn || U.f & 131072) && He() && U.f & 4325394 && (wn === null || !wn.has(e)) && Ie(), Nt(e, n ? F(t) : t, gt);
}
var jt = null, Mt = 0;
function Nt(e, t, n = null) {
	if (!e.equals(t)) {
		bn ? Dt.set(e, t) : Dt.has(e) || Dt.set(e, e.v);
		var r = yt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && ot(t), M === null && qe(t);
		}
		e.wv = jn(), jt = null, Mt = 0, It(e, v, n), jt = null, He() && G !== null && G.f & 1024 && !(G.f & 96) && (J === null ? En([e]) : J.push(e)), !r.is_fork && Et.size > 0 && !Ot && Pt();
	}
	return t;
}
function Pt() {
	Ot = !1;
	for (let e of Et) {
		e.f & 1024 && A(e, y);
		let t;
		try {
			t = Mn(e);
		} catch {
			t = !0;
		}
		t && Rn(e);
	}
	Et.clear();
}
function Ft(e) {
	P(e, e.v + 1);
}
function It(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = He(), a = r.length;
		if (Mt += a, Mt > 1e5 && jt === null && (jt = /* @__PURE__ */ new Set()), jt !== null) {
			if (jt.has(e)) return;
			jt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== G) {
				var l = (c & v) === 0;
				if (l && A(s, t), c & 131072) Et.add(s);
				else if (c & 2) {
					var u = s;
					M?.delete(u), It(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && xt !== null && xt.add(d), n === null ? Ct(d) : n.push(d);
				}
			}
		}
	}
}
function F(e) {
	if (typeof e != "object" || !e || oe in e || se in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ N(0), s = null, l = kn, p = (e) => {
		if (kn === l) return e();
		var t = U, n = kn;
		W(null), An(l);
		var r = e();
		return W(t), An(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ N(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Pe();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ N(n.value, s);
				return i.set(t, e), e;
			}) : P(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ N(t, s));
					i.set(n, e), Ft(o);
				}
			} else P(r, t), Ft(o);
			return !0;
		},
		get(n, r, a) {
			if (r === oe) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ N(F(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = Y(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = Y(a);
				if (o === t) return;
				if (r && "value" in r) r.value = o;
				else return {
					enumerable: !0,
					configurable: !0,
					value: o,
					writable: !0
				};
			}
			return r;
		},
		has(e, n) {
			if (n === oe) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (G !== null || U !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ N(a ? F(e[n]) : t, s)), i.set(n, r)), Y(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ N(t, s)), i.set(f + "", m)) : P(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ N(void 0, s)), P(u, F(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => F(r));
				P(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && P(_, v + 1);
				}
				Ft(o);
			}
			return !0;
		},
		ownKeys(e) {
			Y(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = i.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [r, a] of i) a.v !== t && !(r in e) && n.push(r);
			return n;
		},
		setPrototypeOf() {
			Fe();
		}
	});
}
var Lt, Rt, zt, Bt;
function Vt() {
	if (Lt === void 0) {
		Lt = window, Rt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		zt = c(t, "firstChild").get, Bt = c(t, "nextSibling").get, p(e) && (e[de] = void 0, e[ue] = null, e[fe] = void 0, e.__e = void 0), p(n) && (n[pe] = void 0);
	}
}
function I(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Ht(e) {
	return zt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Ut(e) {
	return Bt.call(e);
}
function L(e, t) {
	if (!T) return /* @__PURE__ */ Ht(e);
	var n = /* @__PURE__ */ Ht(E);
	if (n === null) n = E.appendChild(I());
	else if (t && n.nodeType !== 3) {
		var r = I();
		return n?.before(r), D(r), r;
	}
	return t && Jt(n), D(n), n;
}
function Wt(e, t = !1) {
	if (!T) {
		var n = /* @__PURE__ */ Ht(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Ut(n) : n;
	}
	if (t) {
		if (E?.nodeType !== 3) {
			var r = I();
			return E?.before(r), D(r), r;
		}
		Jt(E);
	}
	return E;
}
function R(e, t = !1) {
	if (!T) return /* @__PURE__ */ Ht(e);
	var n = L(e, t);
	return O(e), n;
}
function z(e, t = 1, n = !1) {
	let r = T ? E : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Ut(r);
	if (!T) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = I();
			return r === null ? i?.after(a) : r.before(a), D(a), a;
		}
		Jt(r);
	}
	return D(r), r;
}
function Gt(e) {
	e.textContent = "";
}
function Kt() {
	return !1;
}
function qt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function Jt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Yt(e) {
	var t = G;
	if (t === null) return U.f |= ae, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	Xt(e, t);
}
function Xt(t, n) {
	if (t === e) throw t;
	if (!(n !== null && n.f & 16384)) {
		for (; n !== null;) {
			if (n.f & 128 && !(n.f & 33570816)) {
				if (!(n.f & 32768)) throw t;
				try {
					n.b.error(t);
					return;
				} catch (e) {
					t = e;
				}
			}
			n = n.parent;
		}
		throw t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/effects.js
function Zt(e) {
	G === null && (U === null && je(e), Ae()), bn && ke(e);
}
function Qt(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function $t(e, t) {
	var n = G;
	n !== null && n.f & 8192 && (e |= b);
	var r = {
		ctx: k,
		deps: null,
		nodes: null,
		f: e | v | 512,
		first: null,
		fn: t,
		last: null,
		next: null,
		parent: n,
		b: n && n.b,
		prev: null,
		teardown: null,
		wv: 0,
		ac: null
	};
	j?.register_created_effect(r);
	var i = r;
	if (e & 4) ht === null ? yt.ensure().schedule(r) : ht.push(r);
	else if (t !== null) {
		try {
			Rn(r);
		} catch (e) {
			throw H(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= C));
	}
	if (i !== null && (i.parent = n, n !== null && Qt(i, n), U !== null && U.f & 2 && !(e & 64))) {
		var a = U;
		(a.effects ??= []).push(i);
	}
	return r;
}
function en() {
	return U !== null && !Sn;
}
function tn(e) {
	let t = $t(8, null);
	return A(t, _), t.teardown = e, t;
}
function nn(e) {
	Zt("$effect");
	var t = G.f;
	if (!U && t & 32 && k !== null && !k.i) {
		var n = k;
		(n.e ??= []).push(e);
	} else return rn(e);
}
function rn(e) {
	return $t(4 | te, e);
}
function an(e) {
	yt.ensure();
	let t = $t(64 | w, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? mn(t, () => {
			H(t), n(void 0);
		}) : (H(t), n(void 0));
	});
}
function on(e) {
	return $t(ie | w, e);
}
function sn(e, t = 0) {
	return $t(8 | t, e);
}
function B(e, t = [], n = [], r = []) {
	Xe(r, t, n, (t) => {
		$t(8, () => {
			e(...t.map(Y));
		});
	});
}
function cn(e, t = 0) {
	return $t(16 | t, e);
}
function V(e) {
	return $t(32 | w, e);
}
function ln(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = bn, r = U;
		xn(!0), W(null);
		try {
			t.call(null);
		} catch (t) {
			Xt(t, e.parent);
		} finally {
			xn(n), W(r);
		}
	}
}
function un(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Ye(() => {
			e.abort(me);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : H(n, t), n = r;
	}
}
function dn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || H(t), t = n;
	}
}
function H(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (fn(e.nodes.start, e.nodes.end), n = !0), e.f |= ee, un(e, t && !n), Ln(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	ln(e), e.f ^= ee, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && pn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function fn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Ut(e);
		e.remove(), e = n;
	}
}
function pn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function mn(e, t, n = !0) {
	var r = [];
	e.f |= 256, hn(e, r, !0);
	var i = () => {
		n && H(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function hn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= b;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				hn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function gn(e) {
	e.f &= -257, _n(e, !0);
}
function _n(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= b, e.f & 1024 || (A(e, v), yt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			_n(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function vn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Ut(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var yn = null, bn = !1;
function xn(e) {
	bn = e;
}
var U = null, Sn = !1;
function W(e) {
	U = e;
}
var G = null;
function Cn(e) {
	G = e;
}
var wn = null;
function Tn(e) {
	U !== null && (U.f & 2097152 || U.f & 2) && (wn ??= /* @__PURE__ */ new Set()).add(e);
}
var K = null, q = 0, J = null;
function En(e) {
	J = e;
}
var Dn = 1, On = 0, kn = On;
function An(e) {
	kn = e;
}
function jn() {
	return ++Dn;
}
function Mn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Mn(a) && st(a), a.wv > e.wv) return !0;
		}
		t & 512 && M === null && A(e, _);
	}
	return !1;
}
function Nn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(wn !== null && wn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Nn(a, t, !1) : t === a && (n ? A(a, v) : a.f & 1024 && A(a, y), Ct(a));
	}
}
function Pn(e) {
	var t = K, n = q, r = J, i = U, a = wn, o = k, s = Sn, c = kn, l = e.f;
	K = null, q = 0, J = null, U = l & 96 ? null : e, wn = null, Re(e.ctx), Sn = !1, kn = ++On, e.ac !== null && (Ye(() => {
		e.ac.abort(me);
	}), e.ac = null);
	try {
		e.f |= re;
		var u = e.fn, d = u();
		e.f |= S;
		var f = Fn(e);
		if (He() && J !== null && !Sn && f !== null && !(e.f & 6146)) for (var p = 0; p < J.length; p++) Nn(J[p], e);
		if (i !== null && i !== e) {
			if (On++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = On;
			if (t !== null) for (let e of t) e.rv = On;
			J !== null && (r === null ? r = J : r.push(...J));
		}
		return e.f & 8388608 && (e.f ^= ae), d;
	} catch (t) {
		return Fn(e), Yt(t);
	} finally {
		e.f ^= re, K = t, q = n, J = r, U = i, wn = a, Re(o), Sn = s, kn = c;
	}
}
function Fn(e) {
	var t = e.deps, n = j?.is_fork;
	if (K !== null) {
		var r;
		if (n || Ln(e, q), t !== null && q > 0) for (t.length = q + K.length, r = 0; r < K.length; r++) t[q + r] = K[r];
		else e.deps = t = K;
		if (en() && e.f & 512) for (r = q; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && q < t.length && (Ln(e, q), t.length = q);
	return t;
}
function In(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (K === null || !a.call(K, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && qe(c), c.ac !== null && Ye(() => {
			c.ac.abort(me), c.ac = null, A(c, v);
		}), ct(c), Ln(c, 0);
	}
}
function Ln(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) In(e, n[r]);
}
function Rn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		A(e, _);
		var n = G;
		G = e;
		try {
			t & 16777232 ? dn(e) : un(e), ln(e);
			var r = Pn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Dn;
		} finally {
			G = n;
		}
	}
}
function Y(e) {
	var t = !!(e.f & 2);
	if (yn?.add(e), U !== null && !Sn && !(G !== null && G.f & 16384) && (wn === null || !wn.has(e))) {
		var n = U.deps;
		if (U.f & 2097152) e.rv < On && (e.rv = On, K === null && n !== null && n[q] === e ? q++ : K === null ? K = [e] : K.push(e));
		else {
			U.deps ??= [], a.call(U.deps, e) || U.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [U] : a.call(r, U) || r.push(U);
		}
	}
	if (bn && Dt.has(e)) return Dt.get(e);
	if (t) {
		var i = e;
		if (bn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Bn(i)) && (o = ot(i)), Dt.set(i, o), o;
		}
		var s = !(i.f & 512) && !Sn && U !== null && !!(U.f & 512), c = (i.f & S) === 0;
		Mn(i) && (s && (i.f |= 512), st(i)), s && !c && (lt(i), zn(i));
	}
	if (M?.has(e)) return M.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function zn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (lt(n), zn(n));
	}
}
function Bn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Dt.has(t) || t.f & 2 && Bn(t)) return !0;
	return !1;
}
function Vn(e) {
	var t = Sn;
	try {
		return Sn = !0, e();
	} finally {
		Sn = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Hn = Symbol("events"), Un = /* @__PURE__ */ new Set(), Wn = /* @__PURE__ */ new Set();
function Gn(e, t, n) {
	(t[Hn] ??= {})[e] = n;
}
function Kn(e) {
	for (var t = 0; t < e.length; t++) Un.add(e[t]);
	for (var n of Wn) n(e);
}
var qn = null, Jn = !1;
function Yn(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	qn = e, Jn || (Jn = !0, setTimeout(() => {
		Jn = !1, qn = null;
	}));
	var o = 0, c = qn === e && e[Hn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Hn] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		l <= u && (o = l);
	}
	if (a = i[o] || e.target, a !== t) {
		s(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = U, f = G;
		W(null), Cn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Hn]?.[r];
					h == null || a.disabled && e.target !== a || h.call(a, e);
				} catch (e) {
					p ? m.push(e) : p = e;
				}
				if (e.cancelBubble) break;
				o++, a = o < i.length ? i[o] : null;
			}
			if (p) {
				for (let e of m) queueMicrotask(() => {
					throw e;
				});
				throw p;
			}
		} finally {
			e[Hn] = t, delete e.currentTarget, W(d), Cn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var Xn = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function Zn(e) {
	return Xn?.createHTML(e) ?? e;
}
function Qn(e) {
	var t = qt("template");
	return t.innerHTML = Zn(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function $n(e, t) {
	var n = G;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function X(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (T) return $n(E, null), E;
		i === void 0 && (i = Qn(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Ht(i)));
		var t = r || Rt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Ht(t), s = t.lastChild;
			$n(o, s);
		} else $n(t, t);
		return t;
	};
}
function er(e = "") {
	if (!T) {
		var t = I(e + "");
		return $n(t, t), t;
	}
	var n = E;
	return n.nodeType === 3 ? Jt(n) : (n.before(n = I()), D(n)), $n(n, n), n;
}
function tr() {
	if (T) return $n(E, null), E;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = I();
	return e.append(t, n), $n(t, n), e;
}
function Z(e, t) {
	if (T) {
		var n = G;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = E), be();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var nr = ["touchstart", "touchmove"];
function rr(e) {
	return nr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function ir(e) {
	let t = 0, n = kt(0), r;
	return () => {
		en() && (Y(n), sn(() => (t === 0 && (r = Vn(() => e(() => Ft(n)))), t += 1, () => {
			Ge(() => {
				--t, t === 0 && (r?.(), r = void 0, Ft(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var ar = C | w;
function or(e, t, n, r) {
	new sr(e, t, n, r);
}
var sr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = T ? E : null;
	#n;
	#r;
	#i;
	#a = null;
	#o = null;
	#s = null;
	#c = null;
	#l = 0;
	#u = 0;
	#d = !1;
	#f = /* @__PURE__ */ new Map();
	#p = null;
	#m = ir(() => (this.#p = kt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = G;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = G.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = cn(() => {
			if (T) {
				let e = Ce(this.#t);
				be();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, ar), T && (this.#e = E);
	}
	#h() {
		try {
			this.#a = V(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Ge(r), t && (this.#s = V(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ve() : (t = !0, n && Le(), this.#s !== null && mn(this.#s, () => {
				this.#s = null;
			}), this.#S(() => {
				this.#b();
			})));
		};
		return {
			reset: r,
			invoke_onerror: () => {
				if (!this.#v()) try {
					n = !0, this.#n.onerror?.(e, r), n = !1;
				} catch (e) {
					Xt(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | ee)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = V(() => e(this.#e)), Ge(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = I(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return V(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							Xt(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(j)) : this.#u === 0 && (this.#e.before(e), this.#c = null, mn(this.#o, () => {
					this.#o = null;
				}), this.#x(j));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = V(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				vn(this.#a, e);
				let t = this.#n.pending;
				this.#o = V(() => t(this.#e));
			} else this.#x(j);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Je(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = G, n = U, r = k;
		Cn(this.#i), W(this.#i), Re(this.#i.ctx);
		try {
			return yt.ensure(), e();
		} finally {
			Cn(t), W(n), Re(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && mn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ge(() => {
			this.#d = !1, this.#p && Nt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), Y(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		j?.is_fork ? (this.#a && j.skip_effect(this.#a), this.#o && j.skip_effect(this.#o), this.#s && j.skip_effect(this.#s), j.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (H(this.#a), null), this.#o &&= (H(this.#o), null), this.#s &&= (H(this.#s), null), T && (D(this.#t), xe(), D(Se()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return V(() => {
						var r = G;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return Xt(e, this.#i.parent), null;
				}
			}));
		};
		Ge(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					Xt(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => Xt(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Q(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[pe] ??= e.nodeValue) && (e[pe] = n, e.nodeValue = `${n}`);
}
function cr(e, t) {
	return ur(e, t);
}
var lr = /* @__PURE__ */ new Map();
function ur(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Vt();
	var u = void 0, d = an(() => {
		var c = r ?? n.appendChild(I());
		or(c, { pending: () => {} }, (n) => {
			ze({});
			var r = k;
			if (s && (r.c = s), a && (i.$$events = a), T && $n(n, null), u = t(n, i) || Ve(), T && (G.nodes.end = E, E === null || E.nodeType !== 8 || E.data !== "]")) throw _e(), e;
			Be();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = rr(r);
					for (let e of [n, document]) {
						var a = lr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), lr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Yn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Un)), Wn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = lr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Yn), t.delete(e), t.size === 0 && lr.delete(r)) : t.set(e, i);
			}
			Wn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return dr.set(u, d), u;
}
var dr = /* @__PURE__ */ new WeakMap();
function fr(e, t) {
	let n = dr.get(e);
	return n ? (dr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var pr = class {
	anchor;
	#e = /* @__PURE__ */ new Map();
	#t = /* @__PURE__ */ new Map();
	#n = /* @__PURE__ */ new Map();
	#r = /* @__PURE__ */ new Set();
	#i = !0;
	constructor(e, t = !0) {
		this.anchor = e, this.#i = t;
	}
	#a = (e) => {
		if (this.#e.has(e)) {
			var t = this.#e.get(e), n = this.#t.get(t);
			if (n) gn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (gn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (H(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						vn(r, t), t.append(I()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else H(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), mn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (H(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = j, r = Kt();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = I();
				i.append(a), this.#n.set(e, {
					effect: V(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, V(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else T && (this.anchor = E), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function $(e, t, n = !1) {
	var r;
	T && (r = E, be());
	var i = new pr(e), a = n ? C : 0;
	function o(e, t) {
		if (T) {
			var n = Ce(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Se();
				D(a), i.anchor = a, ye(!1), i.ensure(e, t), ye(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	cn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function mr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		mn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					hr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Gt(d), d.append(u), e.items.clear();
		}
		hr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function hr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= ne, vn(a, document.createDocumentFragment())) : H(t[i], n);
	}
}
var gr;
function _r(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = T ? D(/* @__PURE__ */ Ht(u)) : u.appendChild(I());
	}
	T && be();
	var d = null, f = /* @__PURE__ */ it(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = Y(f);
			_.fallback = d, yr(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ne, xr(d, null, c)) : gn(d) : mn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: cn(() => {
			var e = Y(f), r = e.length;
			let o = !1;
			T && Ce(c) === "[!" != (r === 0) && (c = Se(), D(c), ye(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = j, v = Kt(), y = 0; y < r; y += 1) {
				T && E.nodeType === 8 && E.data === "]" && (c = E, o = !0, ye(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && Nt(S.v, b), S.i && Nt(S.i, y), v && _.unskip_effect(S.e)) : (S = br(l, m ? c : gr ??= I(), b, x, y, a, t, n), m || (S.e.f |= ne), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = V(() => s(c)) : (d = V(() => s(gr ??= I())), d.f |= ne)), r > u.size && Oe("", "", ""), T && r > 0 && D(Se()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && ye(!0), Y(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, T && (c = E);
}
function vr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function yr(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = vr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (gn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= ne, _ === l) xr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), Sr(e, d, _), Sr(e, _, y), xr(_, y, n), d = _, p = [], m = [], l = vr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) xr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					Sr(e, S.prev, ee.next), Sr(e, d, S), Sr(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), xr(_, l, n), Sr(e, _.prev, _.next), Sr(e, _, d === null ? e.effect.first : d.next), Sr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = vr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = vr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (hr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var C = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || C.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && C.push(l), l = vr(l.next);
		var w = C.length;
		if (w > 0) {
			var te = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < w; v += 1) C[v].nodes?.a?.measure();
				for (v = 0; v < w; v += 1) C[v].nodes?.a?.fix();
			}
			mr(e, C, te);
		}
	}
	a && Ge(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function br(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? kt(n) : /* @__PURE__ */ At(n, !1, !1) : null, l = o & 2 ? kt(i) : null;
	return {
		v: c,
		i: l,
		e: V(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function xr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Ut(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function Sr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var Cr = [..." 	\n\r\f\xA0\v﻿"];
function wr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Cr.includes(r[o - 1])) && (s === r.length || Cr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Tr(e, t, n, r, i, a) {
	var o = e[de];
	if (T || o !== n || o === void 0) {
		var s = wr(n, r, a);
		T && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[de] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Er = Symbol("is custom element"), Dr = Symbol("is html"), Or = he ? "link" : "LINK";
function kr(e, t, n, r) {
	var i = Ar(e);
	if (!(T && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Or))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[le] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Mr(e).has(t) ? e[t] = n : (!T || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Ar(e) {
	return e[ue] ??= {
		[Er]: e.nodeName.includes("-"),
		[Dr]: e.namespaceURI === n
	};
}
var jr = /* @__PURE__ */ new Map();
function Mr(e) {
	var t = e.getAttribute("is") || e.nodeName, n = jr.get(t);
	if (n) return n;
	jr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Nr = !1;
function Pr(e) {
	var t = Nr;
	try {
		return Nr = !1, [e(), Nr];
	} finally {
		Nr = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
function Fr(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, l = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ et(r), Y(u)) : (l && (l = !1, s = o ? Vn(r) : r), s);
	let f;
	if (a) {
		var p = oe in e || ce in e;
		f = c(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Pr(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Ne(t), f(m)));
	var g = i ? () => {
		var n = e[t];
		return n === void 0 ? d() : (l = !0, n);
	} : () => {
		var n = e[t];
		return n !== void 0 && (s = void 0), n === void 0 ? s : n;
	};
	if (i && !(n & 4)) return g;
	if (f) {
		var _ = e.$$legacy;
		return (function(e, t) {
			return arguments.length > 0 ? ((!i || !t || _ || h) && f(t ? g() : e), e) : g();
		});
	}
	var v = !1, y = (n & 1 ? et : it)(() => (v = !1, g()));
	a && Y(y);
	var b = G;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? Y(y) : i && a ? F(e) : e;
			return P(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return bn && v || b.f & 16384 ? y.v : Y(y);
	});
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/attention.js
function Ir(e, t, n) {
	if (!t) return [];
	let r = (t, n) => () => e.call(t, n), i = () => e.post("api/attention/ack", {
		kind: t,
		id: n
	});
	switch (t) {
		case "run": return [{
			label: "Resume run",
			run: r("run.resume", { id: n }),
			ask: "Resume run " + n + "?"
		}];
		case "pool": return [{
			label: "Unquarantine pool",
			run: r("pool.unquarantine", { pool: n }),
			ask: "Unquarantine pool " + n + "?"
		}];
		case "wait": return [{
			label: "Clear wait",
			run: r("wait.clear", {
				id: n,
				reason: "cleared in Desk"
			}),
			ask: "Clear the wait on " + n + "?"
		}];
		case "decision": return [
			{
				label: "Approve as target",
				run: r("decision.resolve", {
					id: n,
					action: "approve"
				}),
				ask: "Approve " + n + ": relabel the issue proposed to target?"
			},
			{
				label: "Reject",
				run: r("decision.resolve", {
					id: n,
					action: "reject"
				}),
				ask: "Reject " + n + ": close the issue as not planned?"
			},
			{
				label: "Dismiss",
				run: r("decision.resolve", {
					id: n,
					action: "dismiss"
				}),
				ask: "Dismiss " + n + " for now? The issue is left as is."
			}
		];
		case "ticket":
		case "question":
		case "stale": return [{
			label: "Close ticket",
			run: r("ticket.close", { id: n }),
			ask: "Close " + n + " as done?"
		}];
		default: return [{
			label: "Acknowledge",
			run: i,
			ask: "Acknowledge " + t + ":" + n + "? It is hidden until it recurs."
		}];
	}
}
async function Lr(e, t) {
	if (!await e.confirm(t.ask)) return !1;
	try {
		return await t.run(), e.toast(t.label + ": done"), !0;
	} catch (n) {
		return e.toast(t.label + " failed: " + String(n.message || n)), !1;
	}
}
var Rr = Symbol.for("jobshop.desk.attention.pending"), zr = "jobshop-attention-pending";
function Br() {
	let e = globalThis;
	return e[Rr] ??= /* @__PURE__ */ new Map();
}
function Vr(e, t) {
	Br().set(e, t), globalThis.dispatchEvent?.(new CustomEvent(zr, { detail: {
		key: e,
		label: t
	} }));
}
function Hr(e) {
	Br().delete(e);
}
function Ur(e) {
	let t = (t) => e(t.detail.key, t.detail.label);
	return globalThis.addEventListener?.(zr, t), () => globalThis.removeEventListener?.(zr, t);
}
// kit/format.js
function Wr(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
// kit/Holder.svelte
var Gr = /* @__PURE__ */ X("<span class=\"svelte-bwb1ob\"> </span>"), Kr = /* @__PURE__ */ X("<span class=\"faint svelte-bwb1ob\" data-testid=\"holder-pool\"> </span>"), qr = /* @__PURE__ */ X("<span class=\"tool muted svelte-bwb1ob\" data-testid=\"holder-tool\"> </span>"), Jr = /* @__PURE__ */ X("<button type=\"button\" class=\"holder svelte-bwb1ob\" data-testid=\"holder\"><span></span> <!> <span class=\"name svelte-bwb1ob\" data-testid=\"holder-name\"> </span> <!> <!></button>");
function Yr(e, t) {
	ze(t, !0);
	let n = Fr(t, "label", 3, "");
	var r = Jr(), i = L(r);
	let a;
	var o = z(i, 2), s = (e) => {
		var t = Gr(), r = R(t, !0);
		B(() => Q(r, n())), Z(e, t);
	};
	$(o, (e) => {
		n() && e(s);
	});
	var c = z(o, 2), l = R(c, !0), u = z(c, 2), d = (e) => {
		var n = Kr(), r = R(n, !0);
		B((e) => Q(r, e), [() => [t.h.pool, t.h.model].filter(Boolean).join(" · ")]), Z(e, n);
	};
	$(u, (e) => {
		(t.h.pool || t.h.model) && e(d);
	});
	var f = z(u, 2), p = (e) => {
		var n = qr(), r = R(n, !0);
		B(() => Q(r, t.h.tool || "working")), Z(e, n);
	};
	$(f, (e) => {
		t.h.busy && e(p);
	}), O(r), B(() => {
		kr(r, "title", "Open " + t.h.id + " read-only"), kr(r, "data-session", t.h.id), a = Tr(i, 1, "dot svelte-bwb1ob", null, a, {
			busy: t.h.busy,
			dead: t.h.alive === !1
		}), Q(l, t.h.display_name || t.h.id);
	}), Gn("click", r, (e) => {
		e.stopPropagation(), t.onopen(t.h.id);
	}), Z(e, r), Be();
}
Kn(["click"]);
// plugins/attention/Attention.svelte
var Xr = /* @__PURE__ */ X("<button class=\"ghost\" data-testid=\"push-test\">Test</button>"), Zr = /* @__PURE__ */ X("<!> <button data-testid=\"push-toggle\"> </button>", 1), Qr = /* @__PURE__ */ X("<span class=\"faint pushna svelte-t4vask\" data-testid=\"push-state\"> </span>"), $r = /* @__PURE__ */ X("<div class=\"card err\"> </div>"), ei = /* @__PURE__ */ X("<p class=\"faint\">Loading…</p>"), ti = /* @__PURE__ */ X("<div class=\"card empty\" data-testid=\"attention-empty\"><div class=\"big\">Nothing needs you.</div> <div class=\"muted\">No escalations, questions, decisions or other items waiting on you.</div></div>"), ni = /* @__PURE__ */ X("<a target=\"_blank\" rel=\"noopener noreferrer\" data-testid=\"attention-ref\"> </a>"), ri = /* @__PURE__ */ X("· <!>", 1), ii = /* @__PURE__ */ X("<div class=\"summary muted svelte-t4vask\" data-testid=\"attention-summary\"> </div>"), ai = /* @__PURE__ */ X("<span class=\"muted\" data-testid=\"row-pending\"> </span>"), oi = /* @__PURE__ */ X("<span class=\"warn-text svelte-t4vask\" data-testid=\"row-stuck\"> </span>"), si = /* @__PURE__ */ X("<button type=\"button\" data-testid=\"row-resolve\" class=\"svelte-t4vask\"> </button>"), ci = /* @__PURE__ */ X("<!> <!> <button type=\"button\" class=\"ghost svelte-t4vask\" data-testid=\"open-chat\">Open chat</button>", 1), li = /* @__PURE__ */ X("<tr data-testid=\"attention-row\"><td class=\"c-id svelte-t4vask\"><span class=\"pill warn\"> </span></td><td class=\"c-title svelte-t4vask\"> <div class=\"faint mono\"> <!></div> <!> <div class=\"acts svelte-t4vask\" role=\"presentation\"><!></div></td><td class=\"muted full svelte-t4vask\"> </td><td class=\"muted c-time svelte-t4vask\"> </td></tr>"), ui = /* @__PURE__ */ X("<div class=\"card scroll\"><table class=\"stack\" data-testid=\"attention-list\"><thead><tr><th>Kind</th><th>Item</th><th>Why</th><th>Since</th></tr></thead><tbody></tbody></table></div>"), di = /* @__PURE__ */ X("<span class=\"step muted svelte-t4vask\"> </span> <!>", 1), fi = /* @__PURE__ */ X("<span class=\"faint\" data-testid=\"no-holder\">waiting for a session</span>"), pi = /* @__PURE__ */ X("<a class=\"mono svelte-t4vask\"> </a>"), mi = /* @__PURE__ */ X("<span class=\"note svelte-t4vask\"> </span>"), hi = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span> </span> <span class=\"step muted svelte-t4vask\"> </span> <!> <!> <span class=\"faint when svelte-t4vask\"> </span></li>"), gi = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span class=\"pill\"> </span><span class=\"step muted svelte-t4vask\"> </span><span class=\"faint\">no report yet</span></li>"), _i = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span class=\"faint\">waiting for a step to start</span></li>"), vi = /* @__PURE__ */ X("<a class=\"svelte-t4vask\"> </a>"), yi = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span class=\"pill\">comment</span><span class=\"step muted mono svelte-t4vask\"><!></span><span class=\"note svelte-t4vask\"> </span><span class=\"faint when svelte-t4vask\"> </span></li>"), bi = /* @__PURE__ */ X("<div class=\"item card svelte-t4vask\" data-testid=\"progress-card\"><button class=\"head svelte-t4vask\"><span class=\"dot accent\"></span> <span class=\"title svelte-t4vask\"> </span> <span class=\"mono faint id svelte-t4vask\"> </span> <span class=\"faint when svelte-t4vask\"> </span></button> <div class=\"who svelte-t4vask\"><!></div> <ul class=\"svelte-t4vask\"><!> <!></ul></div>"), xi = /* @__PURE__ */ X("<div class=\"progress svelte-t4vask\" data-testid=\"progress\"></div>"), Si = /* @__PURE__ */ X("<div class=\"faint\">Loading…</div>"), Ci = /* @__PURE__ */ X("<div class=\"faint none svelte-t4vask\" data-testid=\"progress-empty\">No runs in progress.</div>"), wi = /* @__PURE__ */ X("<header class=\"vh\"><h1>Attention</h1> <span class=\"sub\">What waits on a human, oldest first.</span> <span class=\"actions\"><!></span></header> <div class=\"pad\"><!> <!> <h2 class=\"st\">In progress</h2> <!></div>", 1);
function Ti(e, t) {
	ze(t, !0);
	let n = /* @__PURE__ */ N(null), r = /* @__PURE__ */ N(""), i = /* @__PURE__ */ N(F([])), a = /* @__PURE__ */ N(F({})), o = /* @__PURE__ */ N(!1), s = /* @__PURE__ */ N(""), c = /* @__PURE__ */ N(F({
		state: "off",
		error: ""
	})), l = F({}), u = F({}), d = 0;
	async function f() {
		let e = Object.keys(l);
		try {
			let i = await t.api.get("api/attention"), a = new Set(i.map((e) => e.kind + ":" + e.id)), o = !1;
			for (let t of e) {
				let e = l[t];
				e && (a.has(t) ? ++e.loads >= 2 ? (u[t] = e.label, delete l[t], Hr(t)) : o = !0 : (delete l[t], Hr(t)));
			}
			P(n, i, !0), o && !d && (d = window.setTimeout(() => {
				d = 0, f();
			}, 1500)), P(r, "");
		} catch (e) {
			P(r, String(e.message || e), !0);
		}
		try {
			let e = await t.api.get("api/work?state=running&limit=6&detail=1"), n = e.detail ?? {};
			P(a, e.sessions ?? {}, !0), P(i, (e.runs ?? []).slice(0, 6).map((e) => ({
				run: e,
				steps: n[e.id]?.steps ?? [],
				comment: n[e.id]?.comment ?? null
			})), !0);
		} catch {}
		P(o, !0);
	}
	function p(e, t) {
		delete u[e], l[e] = {
			label: t,
			loads: 0
		};
	}
	nn(() => Vn(() => {
		for (let [e, t] of Br()) p(e, t);
		let e = Ur((e, t) => {
			p(e, t), f();
		});
		f(), t.api.push.status().then((e) => P(c, e, !0));
		let n = 0, r = t.api.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, f();
			}, 1e3);
		}), i = t.api.params();
		if (i.item && !i.session) {
			let [e, ...t] = i.item.split(":");
			m({
				kind: e,
				id: t.join(":")
			});
		}
		return () => {
			r(), e(), clearTimeout(n), clearTimeout(d), d = 0;
		};
	}));
	async function m(e) {
		let n = e.kind + ":" + e.id;
		P(s, n);
		try {
			let r = await t.api.post("api/attention/chat", {
				kind: e.kind,
				id: e.id,
				title: e.title || "",
				why: e.why || ""
			});
			t.api.openSession(r.session, { item: n });
		} catch (r) {
			t.api.toast("Could not open a chat for " + n + ": " + String(r.message || r)), [
				"run",
				"ticket",
				"question",
				"decision",
				"stale",
				"wait"
			].includes(e.kind) && t.api.openTicket(e.id);
		} finally {
			P(s, "");
		}
	}
	let h = /* @__PURE__ */ N("");
	async function g(e, n) {
		let r = e.kind + ":" + e.id;
		P(h, r);
		try {
			await Lr(t.api, n) && Vr(r, n.label);
		} finally {
			P(h, "");
		}
	}
	let _ = {
		unsupported: "This browser has no Web Push (on iOS, add Desk to the home screen first).",
		unavailable: "Desk runs without --push.",
		denied: "Notifications are blocked in the browser settings.",
		off: "Get a notification for each new attention item.",
		on: "Notifications are on for this browser."
	};
	async function v() {
		P(c, Y(c).state === "on" ? await t.api.push.unsubscribe() : await t.api.push.subscribe(), !0);
	}
	async function y() {
		P(c, await t.api.push.test(), !0), Y(c).error || t.api.toast("Test notification sent");
	}
	function b(e) {
		if (!e || e.startsWith("0001")) return "";
		let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3);
		return t < 60 ? t + "s ago" : t < 3600 ? Math.round(t / 60) + "m ago" : t < 86400 ? Math.round(t / 3600) + "h ago" : Math.round(t / 86400) + "d ago";
	}
	let x = (e, t) => typeof e?.[t] == "string" ? JSON.parse(e[t] || "{}") : e?.[t] || {};
	function S(e) {
		return e.map((e) => ({
			s: e,
			r: x(e, "meta").report
		})).filter((e) => e.r && (e.r.note || e.r.state)).sort((e, t) => String(t.r.at).localeCompare(String(e.r.at)));
	}
	let ee = (e) => e === "failed" || e === "blocked" ? "bad" : e === "waiting" || e === "dropped" ? "warn" : e === "ready" ? "ok" : "accent", C = (e) => !!e && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(e), w = (e) => e?.status === "in_progress" && e.holder ? Y(a)[e.holder] ?? { id: e.holder } : null, te = (e) => t.api.openSession(e, { readonly: !0 });
	var ne = wi(), re = Wt(ne), ie = z(L(re), 4), ae = L(ie), oe = (e) => {
		var t = Zr(), n = Wt(t), r = (e) => {
			var t = Xr();
			Gn("click", t, y), Z(e, t);
		};
		$(n, (e) => {
			Y(c).state === "on" && e(r);
		});
		var i = z(n, 2);
		let a;
		var o = R(i, !0);
		B(() => {
			kr(i, "title", _[Y(c).state]), a = Tr(i, 1, "", null, a, { on: Y(c).state === "on" }), Q(o, Y(c).state === "on" ? "Notifications on" : "Notify me");
		}), Gn("click", i, v), Z(e, t);
	}, se = (e) => {
		var t = Qr(), n = R(t, !0);
		B(() => {
			kr(t, "title", _[Y(c).state]), Q(n, Y(c).state === "unavailable" ? "" : "no notifications");
		}), Z(e, t);
	};
	$(ae, (e) => {
		Y(c).state === "on" || Y(c).state === "off" ? e(oe) : e(se, -1);
	}), O(ie), O(re);
	var ce = z(re, 2), le = L(ce), ue = (e) => {
		var t = $r(), n = R(t, !0);
		B(() => Q(n, Y(c).error)), Z(e, t);
	};
	$(le, (e) => {
		Y(c).error && e(ue);
	});
	var de = z(le, 2), fe = (e) => {
		var t = $r(), n = R(t, !0);
		B(() => Q(n, Y(r))), Z(e, t);
	}, pe = (e) => {
		Z(e, ei());
	}, me = (e) => {
		Z(e, ti());
	}, he = (e) => {
		var r = ui(), i = L(r), a = z(L(i));
		_r(a, 21, () => Y(n), (e) => e.kind + e.id, (e, n) => {
			let r = /* @__PURE__ */ rt(() => Y(n).kind + ":" + Y(n).id), i = /* @__PURE__ */ rt(() => l[Y(r)]);
			var a = li();
			let o;
			var c = L(a), d = R(L(c), !0);
			O(c);
			var f = z(c), p = L(f, !0), _ = z(p), v = L(_, !0), y = z(v), x = (e) => {
				var t = ri(), r = z(Wt(t)), i = (e) => {
					var t = ni(), r = R(t, !0);
					B(() => {
						kr(t, "href", Y(n).url), Q(r, Y(n).ref);
					}), Gn("click", t, (e) => e.stopPropagation()), Z(e, t);
				}, a = /* @__PURE__ */ rt(() => Y(n).url && Wr(Y(n).url)), o = (e) => {
					var t = er();
					B(() => Q(t, Y(n).ref)), Z(e, t);
				};
				$(r, (e) => {
					Y(a) ? e(i) : e(o, -1);
				}), Z(e, t);
			};
			$(y, (e) => {
				Y(n).ref && e(x);
			}), O(_);
			var S = z(_, 2), ee = (e) => {
				var t = ii(), r = R(t, !0);
				B(() => Q(r, Y(n).summary)), Z(e, t);
			};
			$(S, (e) => {
				Y(n).summary && e(ee);
			});
			var C = z(S, 2), w = L(C), te = (e) => {
				var t = ai(), n = R(t);
				B(() => Q(n, `${Y(i).label ?? ""}…`)), Z(e, t);
			}, ne = (e) => {
				var i = ci(), a = Wt(i), o = (e) => {
					var t = oi(), n = R(t);
					B(() => Q(n, `still listed after ${u[Y(r)] ?? ""}`)), Z(e, t);
				};
				$(a, (e) => {
					u[Y(r)] && e(o);
				});
				var c = z(a, 2);
				_r(c, 17, () => Ir(t.api, Y(n).kind, Y(n).id), (e) => e.label, (e, t) => {
					var i = si(), a = R(i, !0);
					B(() => {
						i.disabled = Y(h) === Y(r), Q(a, Y(t).label);
					}), Gn("click", i, () => g(Y(n), Y(t))), Z(e, i);
				});
				var l = z(c, 2);
				B(() => l.disabled = Y(s) === Y(r)), Gn("click", l, () => m(Y(n))), Z(e, i);
			};
			$(w, (e) => {
				Y(i) ? e(te) : e(ne, -1);
			}), O(C), O(f);
			var re = z(f), ie = R(re, !0), ae = R(z(re), !0);
			O(a), B((e) => {
				o = Tr(a, 1, "click svelte-t4vask", null, o, {
					busy: Y(s) === Y(r) || !!Y(i),
					pending: !!Y(i)
				}), kr(a, "data-item", Y(n).kind + ":" + Y(n).id), Q(d, Y(n).kind), Q(p, Y(n).title || Y(n).id), Q(v, Y(n).id), Q(ie, Y(n).why), Q(ae, e);
			}, [() => b(Y(n).since)]), Gn("click", a, () => m(Y(n))), Gn("click", C, (e) => e.stopPropagation()), Z(e, a);
		}), O(a), O(i), O(r), Z(e, r);
	};
	$(de, (e) => {
		Y(r) ? e(fe) : Y(n) ? Y(n).length === 0 ? e(me, 2) : e(he, -1) : e(pe, 1);
	});
	var ge = z(de, 4), _e = (e) => {
		var n = xi();
		_r(n, 21, () => Y(i), ({ run: e, steps: t, comment: n }) => e.id, (e, n) => {
			let r = () => Y(n).run, i = () => Y(n).steps, a = () => Y(n).comment, o = /* @__PURE__ */ rt(() => S(i())), s = /* @__PURE__ */ rt(() => i().find((e) => e.status === "in_progress")), c = /* @__PURE__ */ rt(() => w(Y(s)));
			var l = bi(), u = L(l), d = z(L(u), 2), f = R(d, !0), p = z(d, 2), m = R(p, !0), h = R(z(p, 2), !0);
			O(u);
			var g = z(u, 2), _ = L(g), v = (e) => {
				var t = di(), n = Wt(t), r = R(n, !0);
				Yr(z(n, 2), {
					get h() {
						return Y(c);
					},
					onopen: te
				}), B((e) => Q(r, e), [() => x(Y(s), "fields").step || Y(s).title]), Z(e, t);
			}, y = (e) => {
				Z(e, fi());
			};
			$(_, (e) => {
				Y(c) ? e(v) : e(y, -1);
			}), O(g);
			var ne = z(g, 2), re = L(ne);
			_r(re, 17, () => Y(o).slice(0, 3), ({ s: e, r: t }) => e.id, (e, t) => {
				let n = () => Y(t).s, r = () => Y(t).r, i = /* @__PURE__ */ rt(() => w(n()));
				var a = hi(), o = L(a), s = R(o, !0), c = z(o, 2), l = R(c, !0), u = z(c, 2), d = (e) => {
					Yr(e, {
						get h() {
							return Y(i);
						},
						onopen: te
					});
				}, f = (e) => {
					var t = pi(), n = R(t, !0);
					B(() => {
						kr(t, "href", "?session=" + r().by), Q(n, r().by);
					}), Gn("click", t, (e) => {
						e.preventDefault(), te(r().by);
					}), Z(e, t);
				}, p = /* @__PURE__ */ rt(() => C(r().by));
				$(u, (e) => {
					Y(i) ? e(d) : Y(p) && e(f, 1);
				});
				var m = z(u, 2), h = (e) => {
					var t = mi(), n = R(t, !0);
					B(() => Q(n, r().note)), Z(e, t);
				};
				$(m, (e) => {
					r().note && e(h);
				});
				var g = R(z(m, 2), !0);
				O(a), B((e, t, n) => {
					Tr(o, 1, `pill ${e ?? ""}`, "svelte-t4vask"), Q(s, r().state), Q(l, t), Q(g, n);
				}, [
					() => ee(r().state),
					() => x(n(), "fields").step || n().title,
					() => b(r().at)
				]), Z(e, a);
			}, (e) => {
				var t = tr(), n = Wt(t), r = (e) => {
					var t = gi(), n = L(t), r = R(n, !0), i = R(z(n), !0);
					xe(), O(t), B((e) => {
						Q(r, Y(s).status), Q(i, e);
					}, [() => x(Y(s), "fields").step || Y(s).title]), Z(e, t);
				}, i = (e) => {
					Z(e, _i());
				};
				$(n, (e) => {
					Y(s) ? e(r) : e(i, -1);
				}), Z(e, t);
			});
			var ie = z(re, 2), ae = (e) => {
				var n = yi(), r = z(L(n)), i = L(r), o = (e) => {
					var n = vi(), r = R(n, !0);
					B(() => {
						kr(n, "href", "?session=" + a().author), Q(r, a().author);
					}), Gn("click", n, (e) => {
						e.preventDefault(), t.api.openSession(a().author);
					}), Z(e, n);
				}, s = /* @__PURE__ */ rt(() => C(a().author)), c = (e) => {
					var t = er();
					B(() => Q(t, a().author)), Z(e, t);
				};
				$(i, (e) => {
					Y(s) ? e(o) : e(c, -1);
				}), O(r);
				var l = z(r), u = R(l, !0), d = R(z(l), !0);
				O(n), B((e) => {
					Q(u, a().body), Q(d, e);
				}, [() => b(a().created_at)]), Z(e, n);
			};
			$(ie, (e) => {
				a() && e(ae);
			}), O(ne), O(l), B((e) => {
				kr(l, "data-run", r().id), Q(f, r().title), Q(m, r().id), Q(h, e);
			}, [() => b(r().updated_at)]), Gn("click", u, () => t.api.openTicket(r().id)), Z(e, l);
		}), O(n), Z(e, n);
	}, ve = (e) => {
		Z(e, Si());
	}, T = (e) => {
		Z(e, Ci());
	};
	$(ge, (e) => {
		Y(i).length ? e(_e) : Y(o) ? e(T, -1) : e(ve, 1);
	}), O(ce), Z(e, ne), Be();
}
Kn(["click"]);
// plugins/attention/index.svelte.js
var Ei = new URL("./style.css", import.meta.url).href, Di = class extends HTMLElement {
	#e = /* @__PURE__ */ N(F({
		api: null,
		context: {},
		config: {}
	}));
	get props() {
		return Y(this.#e);
	}
	set props(e) {
		P(this.#e, e, !0);
	}
	app = null;
	set api(e) {
		this.props.api = e;
	}
	get api() {
		return this.props.api;
	}
	set context(e) {
		this.props.context = e;
	}
	get context() {
		return this.props.context;
	}
	set config(e) {
		this.props.config = e;
	}
	get config() {
		return this.props.config;
	}
	connectedCallback() {
		if (!document.querySelector(`link[href="${Ei}"]`)) {
			let e = document.createElement("link");
			e.rel = "stylesheet", e.href = Ei, document.head.append(e);
		}
		this.style.display = "contents", this.app = cr(Ti, {
			target: this,
			props: this.props
		});
	}
	disconnectedCallback() {
		this.app && fr(this.app), this.app = null;
	}
};
customElements.get("desk-attention-view") || customElements.define("desk-attention-view", Di);
