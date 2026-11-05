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
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, ee = 1 << 25, C = 65536, w = 1 << 19, te = 1 << 20, ne = 1 << 25, re = 1 << 21, ie = 1 << 22, ae = 1 << 23, oe = Symbol("$state"), se = Symbol("component"), ce = Symbol(""), le = Symbol("attributes"), ue = Symbol("class"), de = Symbol("style"), fe = Symbol("text"), pe = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), me = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function he() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function ge(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function _e() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var T = !1;
function ve(e) {
	T = e;
}
var E;
function D(t) {
	if (t === null) throw ge(), e;
	return E = t;
}
function ye() {
	return D(/* @__PURE__ */ Ht(E));
}
function O(t) {
	if (T) {
		if (/* @__PURE__ */ Ht(E) !== null) throw ge(), e;
		E = t;
	}
}
function be(e = 1) {
	if (T) {
		for (var t = e, n = E; t--;) n = /* @__PURE__ */ Ht(n);
		E = n;
	}
}
function xe(e = !0) {
	for (var t = 0, n = E;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Ht(n);
		e && n.remove(), n = i;
	}
}
function Se(t) {
	if (!t || t.nodeType !== 8) throw ge(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Ce(e) {
	return e === this.v;
}
function we(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Te(e) {
	return !we(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function Ee() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function De(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function Oe(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function ke() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Ae(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function je() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Me() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Ne() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Pe() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Fe() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var k = null;
function Ie(e) {
	k = e;
}
function Le(e, t = !1, n) {
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
function Re(e) {
	var t = k, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) nn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, k = t.p, ze(e);
}
function ze(e = {}) {
	return s(e, se, { value: !0 }), e;
}
function Be() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Ve = [];
function He() {
	var e = Ve;
	Ve = [], h(e);
}
function Ue(e) {
	if (Ve.length === 0 && !dt) {
		var t = Ve;
		queueMicrotask(() => {
			t === Ve && He();
		});
	}
	Ve.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var We = ~(v | y | _);
function A(e, t) {
	e.f = e.f & We | t;
}
function Ge(e) {
	e.f & 512 || e.deps === null ? A(e, _) : A(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Ke(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), A(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function qe(e) {
	var t = H, n = G;
	W(null), xn(null);
	try {
		return e();
	} finally {
		W(t), xn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Je(e, t, n, r) {
	let i = Be() ? Qe : nt;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = G, c = Ye(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				Yt(e, s);
			}
			Xe();
		}
	}
	var d = Ze();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ et(e))).then(u).catch((e) => Yt(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), Xe());
	}) : f();
}
function Ye() {
	var e = G, t = H, n = k, r = j;
	return function(i = !0) {
		xn(e), W(t), Ie(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function Xe(e = !0) {
	xn(null), W(null), Ie(null), e && j?.deactivate();
}
function Ze() {
	var e = G, t = e.b, n = j, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Qe(e) {
	var n = 2 | v;
	return G !== null && (G.f |= w), {
		ctx: k,
		deps: null,
		effects: null,
		equals: Ce,
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
var $e = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function et(e, n, r) {
	let i = G;
	i === null && Ee();
	var a = void 0, o = Dt(t), s = !H, c = /* @__PURE__ */ new Set();
	return an(() => {
		var t = G, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== pe && n.reject(e);
			}).finally(Xe);
		} catch (e) {
			n.reject(e), Xe();
		}
		var r = j;
		if (s) {
			if (t.f & 32768) var l = Ze();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject($e);
			else for (let e of c.values()) e.reject($e);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== $e && (r.activate(), t ? (o.f |= ae, jt(o, t)) : (o.f & 8388608 && (o.f ^= ae), jt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), en(() => {
		for (let e of c) e.reject($e);
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
function tt(e) {
	let t = /* @__PURE__ */ Qe(e);
	return Cn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function nt(e) {
	let t = /* @__PURE__ */ Qe(e);
	return t.equals = Te, t;
}
function rt(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) V(t[n]);
	}
}
function it(e) {
	var n, r = G, i = e.parent;
	if (!yn && i !== null && e.v !== t && i.f & 24576) return he(), e.v;
	xn(i);
	try {
		rt(e), n = Mn(e);
	} finally {
		xn(r);
	}
	return n;
}
function at(e) {
	var t = it(e);
	!e.equals(t) && (e.wv = kn(), (!j?.is_fork || e.deps === null) && (j === null ? e.v = t : (j.capture(e, t, !0), lt?.capture(e, t, !0)), e.deps === null)) ? A(e, _) : yn || (M === null ? Ge(e) : ($t() || j?.is_fork) && M.set(e, t));
}
function ot(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && qe(() => {
		t.ac.abort(pe), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Fn(t, 0), ln(t));
}
function st(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && In(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var ct = null, j = null, lt = null, M = null, ut = null, dt = !1, ft = !1, pt = null, mt = null, ht = 0, gt = 1, _t = class e {
	id = gt++;
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
		ct === null ? ct = this : (ct.next = this, this.prev = ct), ct = this;
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
		for (var t = pt = [], n = [], r = mt = []; this.#a.length > 0;) {
			ht++ > 1e3 && (this.#b(), vt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Ct(e), this.#d() || this.discard(), t;
			}
		}
		if (j = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (pt = null, mt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) St(e, t);
			r.length > 0 && j.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), lt = this, bt(n), bt(t), lt = null, this.#i?.resolve();
			var o = j;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Tt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : An(r) && (i & 16 && this.add_dirty_reaction(r, y), In(r));
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
		for (var t = 0; t < e.length; t += 1) Ke(e[t], this.#s);
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
			ft = !0, j = this, this.#m();
		} finally {
			ht = 0, ut = null, pt = null, mt = null, ft = !1, j = null, M = null, Tt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject($e);
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
		this.#u || (this.#u = !0, Ue(() => {
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
			!ft && Ue(() => {
				t.started || t.flush();
			});
		}
		return j;
	}
	apply() {
		M = null;
	}
	schedule(e) {
		ut = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? ct = e : t.prev = e, this.linked = !1;
		}
	}
};
function vt() {
	try {
		je();
	} catch (e) {
		Yt(e, ut);
	}
}
var yt = null;
function bt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && An(r) && (yt = /* @__PURE__ */ new Set(), In(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && fn(r), yt?.size > 0)) {
				Tt.clear();
				for (let e of yt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) yt.has(n) && (yt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || In(n);
					}
				}
				yt.clear();
			}
		}
		yt = null;
	}
}
function xt(e) {
	j.schedule(e);
}
function St(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), A(e, _);
		for (var n = e.first; n !== null;) St(n, t), n = n.next;
	}
}
function Ct(e) {
	A(e, _);
	for (var t = e.first; t !== null;) Ct(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var wt = /* @__PURE__ */ new Set(), Tt = /* @__PURE__ */ new Map(), Et = !1;
function Dt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Ce,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function N(e, t) {
	let n = Dt(e, t);
	return Cn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Ot(e, t = !1, n = !0) {
	let r = Dt(e);
	return t || (r.equals = Te), r;
}
function P(e, t, n = !1) {
	return H !== null && (!U || H.f & 131072) && Be() && H.f & 4325394 && (Sn === null || !Sn.has(e)) && Pe(), jt(e, n ? Ft(t) : t, mt);
}
var kt = null, At = 0;
function jt(e, t, n = null) {
	if (!e.equals(t)) {
		yn ? Tt.set(e, t) : Tt.has(e) || Tt.set(e, e.v);
		var r = _t.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && it(t), M === null && Ge(t);
		}
		e.wv = kn(), kt = null, At = 0, Pt(e, v, n), kt = null, Be() && G !== null && G.f & 1024 && !(G.f & 96) && (J === null ? wn([e]) : J.push(e)), !r.is_fork && wt.size > 0 && !Et && Mt();
	}
	return t;
}
function Mt() {
	Et = !1;
	for (let e of wt) {
		e.f & 1024 && A(e, y);
		let t;
		try {
			t = An(e);
		} catch {
			t = !0;
		}
		t && In(e);
	}
	wt.clear();
}
function Nt(e) {
	P(e, e.v + 1);
}
function Pt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Be(), a = r.length;
		if (At += a, At > 1e5 && kt === null && (kt = /* @__PURE__ */ new Set()), kt !== null) {
			if (kt.has(e)) return;
			kt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== G) {
				var l = (c & v) === 0;
				if (l && A(s, t), c & 131072) wt.add(s);
				else if (c & 2) {
					var u = s;
					M?.delete(u), Pt(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && yt !== null && yt.add(d), n === null ? xt(d) : n.push(d);
				}
			}
		}
	}
}
function Ft(e) {
	if (typeof e != "object" || !e || oe in e || se in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ N(0), s = null, l = Dn, p = (e) => {
		if (Dn === l) return e();
		var t = H, n = Dn;
		W(null), On(l);
		var r = e();
		return W(t), On(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ N(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Me();
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
					i.set(n, e), Nt(o);
				}
			} else P(r, t), Nt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === oe) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ N(Ft(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
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
			return (r !== void 0 || (G !== null || H !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ N(a ? Ft(e[n]) : t, s)), i.set(n, r)), Y(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ N(t, s)), i.set(f + "", m)) : P(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ N(void 0, s)), P(u, Ft(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => Ft(r));
				P(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && P(_, v + 1);
				}
				Nt(o);
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
			Ne();
		}
	});
}
var It, Lt, Rt, zt;
function Bt() {
	if (It === void 0) {
		It = window, Lt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Rt = c(t, "firstChild").get, zt = c(t, "nextSibling").get, p(e) && (e[ue] = void 0, e[le] = null, e[de] = void 0, e.__e = void 0), p(n) && (n[fe] = void 0);
	}
}
function F(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Vt(e) {
	return Rt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Ht(e) {
	return zt.call(e);
}
function I(e, t) {
	if (!T) return /* @__PURE__ */ Vt(e);
	var n = /* @__PURE__ */ Vt(E);
	if (n === null) n = E.appendChild(F());
	else if (t && n.nodeType !== 3) {
		var r = F();
		return n?.before(r), D(r), r;
	}
	return t && qt(n), D(n), n;
}
function Ut(e, t = !1) {
	if (!T) {
		var n = /* @__PURE__ */ Vt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Ht(n) : n;
	}
	if (t) {
		if (E?.nodeType !== 3) {
			var r = F();
			return E?.before(r), D(r), r;
		}
		qt(E);
	}
	return E;
}
function L(e, t = !1) {
	if (!T) return /* @__PURE__ */ Vt(e);
	var n = I(e, t);
	return O(e), n;
}
function R(e, t = 1, n = !1) {
	let r = T ? E : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Ht(r);
	if (!T) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = F();
			return r === null ? i?.after(a) : r.before(a), D(a), a;
		}
		qt(r);
	}
	return D(r), r;
}
function Wt(e) {
	e.textContent = "";
}
function Gt() {
	return !1;
}
function Kt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function qt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Jt(e) {
	var t = G;
	if (t === null) return H.f |= ae, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	Yt(e, t);
}
function Yt(t, n) {
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
function Xt(e) {
	G === null && (H === null && Ae(e), ke()), yn && Oe(e);
}
function Zt(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function Qt(e, t) {
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
	if (e & 4) pt === null ? _t.ensure().schedule(r) : pt.push(r);
	else if (t !== null) {
		try {
			In(r);
		} catch (e) {
			throw V(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= C));
	}
	if (i !== null && (i.parent = n, n !== null && Zt(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function $t() {
	return H !== null && !U;
}
function en(e) {
	let t = Qt(8, null);
	return A(t, _), t.teardown = e, t;
}
function tn(e) {
	Xt("$effect");
	var t = G.f;
	if (!H && t & 32 && k !== null && !k.i) {
		var n = k;
		(n.e ??= []).push(e);
	} else return nn(e);
}
function nn(e) {
	return Qt(4 | te, e);
}
function rn(e) {
	_t.ensure();
	let t = Qt(64 | w, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? pn(t, () => {
			V(t), n(void 0);
		}) : (V(t), n(void 0));
	});
}
function an(e) {
	return Qt(ie | w, e);
}
function on(e, t = 0) {
	return Qt(8 | t, e);
}
function z(e, t = [], n = [], r = []) {
	Je(r, t, n, (t) => {
		Qt(8, () => {
			e(...t.map(Y));
		});
	});
}
function sn(e, t = 0) {
	return Qt(16 | t, e);
}
function B(e) {
	return Qt(32 | w, e);
}
function cn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = yn, r = H;
		bn(!0), W(null);
		try {
			t.call(null);
		} catch (t) {
			Yt(t, e.parent);
		} finally {
			bn(n), W(r);
		}
	}
}
function ln(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && qe(() => {
			e.abort(pe);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : V(n, t), n = r;
	}
}
function un(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || V(t), t = n;
	}
}
function V(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (dn(e.nodes.start, e.nodes.end), n = !0), e.f |= ee, ln(e, t && !n), Fn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	cn(e), e.f ^= ee, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && fn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function dn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Ht(e);
		e.remove(), e = n;
	}
}
function fn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function pn(e, t, n = !0) {
	var r = [];
	e.f |= 256, mn(e, r, !0);
	var i = () => {
		n && V(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function mn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= b;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				mn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function hn(e) {
	e.f &= -257, gn(e, !0);
}
function gn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= b, e.f & 1024 || (A(e, v), _t.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			gn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function _n(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Ht(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var vn = null, yn = !1;
function bn(e) {
	yn = e;
}
var H = null, U = !1;
function W(e) {
	H = e;
}
var G = null;
function xn(e) {
	G = e;
}
var Sn = null;
function Cn(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (Sn ??= /* @__PURE__ */ new Set()).add(e);
}
var K = null, q = 0, J = null;
function wn(e) {
	J = e;
}
var Tn = 1, En = 0, Dn = En;
function On(e) {
	Dn = e;
}
function kn() {
	return ++Tn;
}
function An(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (An(a) && at(a), a.wv > e.wv) return !0;
		}
		t & 512 && M === null && A(e, _);
	}
	return !1;
}
function jn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Sn !== null && Sn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? jn(a, t, !1) : t === a && (n ? A(a, v) : a.f & 1024 && A(a, y), xt(a));
	}
}
function Mn(e) {
	var t = K, n = q, r = J, i = H, a = Sn, o = k, s = U, c = Dn, l = e.f;
	K = null, q = 0, J = null, H = l & 96 ? null : e, Sn = null, Ie(e.ctx), U = !1, Dn = ++En, e.ac !== null && (qe(() => {
		e.ac.abort(pe);
	}), e.ac = null);
	try {
		e.f |= re;
		var u = e.fn, d = u();
		e.f |= S;
		var f = Nn(e);
		if (Be() && J !== null && !U && f !== null && !(e.f & 6146)) for (var p = 0; p < J.length; p++) jn(J[p], e);
		if (i !== null && i !== e) {
			if (En++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = En;
			if (t !== null) for (let e of t) e.rv = En;
			J !== null && (r === null ? r = J : r.push(...J));
		}
		return e.f & 8388608 && (e.f ^= ae), d;
	} catch (t) {
		return Nn(e), Jt(t);
	} finally {
		e.f ^= re, K = t, q = n, J = r, H = i, Sn = a, Ie(o), U = s, Dn = c;
	}
}
function Nn(e) {
	var t = e.deps, n = j?.is_fork;
	if (K !== null) {
		var r;
		if (n || Fn(e, q), t !== null && q > 0) for (t.length = q + K.length, r = 0; r < K.length; r++) t[q + r] = K[r];
		else e.deps = t = K;
		if ($t() && e.f & 512) for (r = q; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && q < t.length && (Fn(e, q), t.length = q);
	return t;
}
function Pn(e, n) {
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
		c.f & 512 && (c.f ^= 512), c.v !== t && Ge(c), c.ac !== null && qe(() => {
			c.ac.abort(pe), c.ac = null, A(c, v);
		}), ot(c), Fn(c, 0);
	}
}
function Fn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Pn(e, n[r]);
}
function In(e) {
	var t = e.f;
	if (!(t & 16384)) {
		A(e, _);
		var n = G;
		G = e;
		try {
			t & 16777232 ? un(e) : ln(e), cn(e);
			var r = Mn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Tn;
		} finally {
			G = n;
		}
	}
}
function Y(e) {
	var t = !!(e.f & 2);
	if (vn?.add(e), H !== null && !U && !(G !== null && G.f & 16384) && (Sn === null || !Sn.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < En && (e.rv = En, K === null && n !== null && n[q] === e ? q++ : K === null ? K = [e] : K.push(e));
		else {
			H.deps ??= [], a.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : a.call(r, H) || r.push(H);
		}
	}
	if (yn && Tt.has(e)) return Tt.get(e);
	if (t) {
		var i = e;
		if (yn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Rn(i)) && (o = it(i)), Tt.set(i, o), o;
		}
		var s = !(i.f & 512) && !U && H !== null && !!(H.f & 512), c = (i.f & S) === 0;
		An(i) && (s && (i.f |= 512), at(i)), s && !c && (st(i), Ln(i));
	}
	if (M?.has(e)) return M.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Ln(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (st(n), Ln(n));
	}
}
function Rn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Tt.has(t) || t.f & 2 && Rn(t)) return !0;
	return !1;
}
function zn(e) {
	var t = U;
	try {
		return U = !0, e();
	} finally {
		U = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Bn = Symbol("events"), Vn = /* @__PURE__ */ new Set(), Hn = /* @__PURE__ */ new Set();
function Un(e, t, n) {
	(t[Bn] ??= {})[e] = n;
}
function Wn(e) {
	for (var t = 0; t < e.length; t++) Vn.add(e[t]);
	for (var n of Hn) n(e);
}
var Gn = null, Kn = !1;
function qn(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	Gn = e, Kn || (Kn = !0, setTimeout(() => {
		Kn = !1, Gn = null;
	}));
	var o = 0, c = Gn === e && e[Bn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Bn] = t;
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
		var d = H, f = G;
		W(null), xn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Bn]?.[r];
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
			e[Bn] = t, delete e.currentTarget, W(d), xn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var Jn = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function Yn(e) {
	return Jn?.createHTML(e) ?? e;
}
function Xn(e) {
	var t = Kt("template");
	return t.innerHTML = Yn(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Zn(e, t) {
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
		if (T) return Zn(E, null), E;
		i === void 0 && (i = Xn(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Vt(i)));
		var t = r || Lt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Vt(t), s = t.lastChild;
			Zn(o, s);
		} else Zn(t, t);
		return t;
	};
}
function Qn(e = "") {
	if (!T) {
		var t = F(e + "");
		return Zn(t, t), t;
	}
	var n = E;
	return n.nodeType === 3 ? qt(n) : (n.before(n = F()), D(n)), Zn(n, n), n;
}
function $n() {
	if (T) return Zn(E, null), E;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = F();
	return e.append(t, n), Zn(t, n), e;
}
function Z(e, t) {
	if (T) {
		var n = G;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = E), ye();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var er = ["touchstart", "touchmove"];
function tr(e) {
	return er.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function nr(e) {
	let t = 0, n = Dt(0), r;
	return () => {
		$t() && (Y(n), on(() => (t === 0 && (r = zn(() => e(() => Nt(n)))), t += 1, () => {
			Ue(() => {
				--t, t === 0 && (r?.(), r = void 0, Nt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var rr = C | w;
function ir(e, t, n, r) {
	new ar(e, t, n, r);
}
var ar = class {
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
	#m = nr(() => (this.#p = Dt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = G;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = G.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = sn(() => {
			if (T) {
				let e = Se(this.#t);
				ye();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, rr), T && (this.#e = E);
	}
	#h() {
		try {
			this.#a = B(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Ue(r), t && (this.#s = B(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? _e() : (t = !0, n && Fe(), this.#s !== null && pn(this.#s, () => {
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
					Yt(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | ee)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = B(() => e(this.#e)), Ue(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = F(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return B(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							Yt(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(j)) : this.#u === 0 && (this.#e.before(e), this.#c = null, pn(this.#o, () => {
					this.#o = null;
				}), this.#x(j));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = B(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				_n(this.#a, e);
				let t = this.#n.pending;
				this.#o = B(() => t(this.#e));
			} else this.#x(j);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Ke(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = G, n = H, r = k;
		xn(this.#i), W(this.#i), Ie(this.#i.ctx);
		try {
			return _t.ensure(), e();
		} finally {
			xn(t), W(n), Ie(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && pn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ue(() => {
			this.#d = !1, this.#p && jt(this.#p, this.#l);
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
		this.#a &&= (V(this.#a), null), this.#o &&= (V(this.#o), null), this.#s &&= (V(this.#s), null), T && (D(this.#t), be(), D(xe()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return B(() => {
						var r = G;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return Yt(e, this.#i.parent), null;
				}
			}));
		};
		Ue(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					Yt(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => Yt(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Q(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[fe] ??= e.nodeValue) && (e[fe] = n, e.nodeValue = `${n}`);
}
function or(e, t) {
	return cr(e, t);
}
var sr = /* @__PURE__ */ new Map();
function cr(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Bt();
	var u = void 0, d = rn(() => {
		var c = r ?? n.appendChild(F());
		ir(c, { pending: () => {} }, (n) => {
			Le({});
			var r = k;
			if (s && (r.c = s), a && (i.$$events = a), T && Zn(n, null), u = t(n, i) || ze(), T && (G.nodes.end = E, E === null || E.nodeType !== 8 || E.data !== "]")) throw ge(), e;
			Re();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = tr(r);
					for (let e of [n, document]) {
						var a = sr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), sr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, qn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Vn)), Hn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = sr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, qn), t.delete(e), t.size === 0 && sr.delete(r)) : t.set(e, i);
			}
			Hn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return lr.set(u, d), u;
}
var lr = /* @__PURE__ */ new WeakMap();
function ur(e, t) {
	let n = lr.get(e);
	return n ? (lr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var dr = class {
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
			if (n) hn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (hn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (V(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						_n(r, t), t.append(F()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else V(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), pn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (V(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = j, r = Gt();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = F();
				i.append(a), this.#n.set(e, {
					effect: B(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, B(() => t(this.anchor)));
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
	T && (r = E, ye());
	var i = new dr(e), a = n ? C : 0;
	function o(e, t) {
		if (T) {
			var n = Se(r);
			if (e !== parseInt(n.substring(1))) {
				var a = xe();
				D(a), i.anchor = a, ve(!1), i.ensure(e, t), ve(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	sn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function fr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		pn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					pr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Wt(d), d.append(u), e.items.clear();
		}
		pr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function pr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= ne, _n(a, document.createDocumentFragment())) : V(t[i], n);
	}
}
var mr;
function hr(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = T ? D(/* @__PURE__ */ Vt(u)) : u.appendChild(F());
	}
	T && ye();
	var d = null, f = /* @__PURE__ */ nt(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = Y(f);
			_.fallback = d, _r(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ne, yr(d, null, c)) : hn(d) : pn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: sn(() => {
			var e = Y(f), r = e.length;
			let o = !1;
			T && Se(c) === "[!" != (r === 0) && (c = xe(), D(c), ve(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = j, v = Gt(), y = 0; y < r; y += 1) {
				T && E.nodeType === 8 && E.data === "]" && (c = E, o = !0, ve(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && jt(S.v, b), S.i && jt(S.i, y), v && _.unskip_effect(S.e)) : (S = vr(l, m ? c : mr ??= F(), b, x, y, a, t, n), m || (S.e.f |= ne), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = B(() => s(c)) : (d = B(() => s(mr ??= F())), d.f |= ne)), r > u.size && De("", "", ""), T && r > 0 && D(xe()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && ve(!0), Y(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, T && (c = E);
}
function gr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function _r(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = gr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (hn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= ne, _ === l) yr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), br(e, d, _), br(e, _, y), yr(_, y, n), d = _, p = [], m = [], l = gr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) yr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					br(e, S.prev, ee.next), br(e, d, S), br(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), yr(_, l, n), br(e, _.prev, _.next), br(e, _, d === null ? e.effect.first : d.next), br(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = gr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = gr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (pr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var C = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || C.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && C.push(l), l = gr(l.next);
		var w = C.length;
		if (w > 0) {
			var te = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < w; v += 1) C[v].nodes?.a?.measure();
				for (v = 0; v < w; v += 1) C[v].nodes?.a?.fix();
			}
			fr(e, C, te);
		}
	}
	a && Ue(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function vr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Dt(n) : /* @__PURE__ */ Ot(n, !1, !1) : null, l = o & 2 ? Dt(i) : null;
	return {
		v: c,
		i: l,
		e: B(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function yr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Ht(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function br(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var xr = [..." 	\n\r\f\xA0\v﻿"];
function Sr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || xr.includes(r[o - 1])) && (s === r.length || xr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Cr(e, t, n, r, i, a) {
	var o = e[ue];
	if (T || o !== n || o === void 0) {
		var s = Sr(n, r, a);
		T && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[ue] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var wr = Symbol("is custom element"), Tr = Symbol("is html"), Er = me ? "link" : "LINK";
function Dr(e, t, n, r) {
	var i = Or(e);
	if (!(T && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Er))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[ce] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Ar(e).has(t) ? e[t] = n : (!T || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Or(e) {
	return e[le] ??= {
		[wr]: e.nodeName.includes("-"),
		[Tr]: e.namespaceURI === n
	};
}
var kr = /* @__PURE__ */ new Map();
function Ar(e) {
	var t = e.getAttribute("is") || e.nodeName, n = kr.get(t);
	if (n) return n;
	kr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// plugins/attention/Attention.svelte
var jr = /* @__PURE__ */ X("<button class=\"ghost\" data-testid=\"push-test\">Test</button>"), Mr = /* @__PURE__ */ X("<!> <button data-testid=\"push-toggle\"> </button>", 1), Nr = /* @__PURE__ */ X("<span class=\"faint pushna svelte-t4vask\" data-testid=\"push-state\"> </span>"), Pr = /* @__PURE__ */ X("<div class=\"card err\"> </div>"), Fr = /* @__PURE__ */ X("<p class=\"faint\">Loading…</p>"), Ir = /* @__PURE__ */ X("<div class=\"card empty\" data-testid=\"attention-empty\"><div class=\"big\">Nothing needs you.</div> <div class=\"muted\">No escalations, questions, decisions, quarantined pools or rolled-back deploys.</div></div>"), Lr = /* @__PURE__ */ X("<tr data-testid=\"attention-row\"><td class=\"c-id svelte-t4vask\"><span class=\"pill warn\"> </span></td><td class=\"c-title svelte-t4vask\"> <div class=\"faint mono\"> <!></div></td><td class=\"muted full svelte-t4vask\"> </td><td class=\"muted c-time svelte-t4vask\"> </td></tr>"), Rr = /* @__PURE__ */ X("<div class=\"card scroll\"><table class=\"stack\" data-testid=\"attention-list\"><thead><tr><th>Kind</th><th>Item</th><th>Why</th><th>Since</th></tr></thead><tbody></tbody></table></div>"), zr = /* @__PURE__ */ X("<a class=\"mono svelte-t4vask\"> </a>"), Br = /* @__PURE__ */ X("<span class=\"note svelte-t4vask\"> </span>"), Vr = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span> </span> <span class=\"step muted svelte-t4vask\"> </span> <!> <!> <span class=\"faint when svelte-t4vask\"> </span></li>"), Hr = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span class=\"pill\"> </span><span class=\"step muted svelte-t4vask\"> </span><span class=\"faint\">no report yet</span></li>"), Ur = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span class=\"faint\">waiting for a step to start</span></li>"), Wr = /* @__PURE__ */ X("<a class=\"svelte-t4vask\"> </a>"), Gr = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span class=\"pill\">comment</span><span class=\"step muted mono svelte-t4vask\"><!></span><span class=\"note svelte-t4vask\"> </span><span class=\"faint when svelte-t4vask\"> </span></li>"), Kr = /* @__PURE__ */ X("<div class=\"item card svelte-t4vask\"><button class=\"head svelte-t4vask\"><span class=\"dot accent\"></span> <span class=\"title svelte-t4vask\"> </span> <span class=\"mono faint id svelte-t4vask\"> </span> <span class=\"faint when svelte-t4vask\"> </span></button> <ul class=\"svelte-t4vask\"><!> <!></ul></div>"), qr = /* @__PURE__ */ X("<div class=\"progress svelte-t4vask\" data-testid=\"progress\"></div>"), Jr = /* @__PURE__ */ X("<div class=\"faint\">Loading…</div>"), Yr = /* @__PURE__ */ X("<div class=\"faint none svelte-t4vask\" data-testid=\"progress-empty\">No runs in progress.</div>"), Xr = /* @__PURE__ */ X("<header class=\"vh\"><h1>Attention</h1> <span class=\"sub\">What waits on a human, oldest first.</span> <span class=\"actions\"><!></span></header> <div class=\"pad\"><!> <!> <h2 class=\"st\">In progress</h2> <!></div>", 1);
function Zr(e, t) {
	Le(t, !0);
	let n = /* @__PURE__ */ N(null), r = /* @__PURE__ */ N(""), i = /* @__PURE__ */ N(Ft([])), a = /* @__PURE__ */ N(!1), o = /* @__PURE__ */ N(""), s = /* @__PURE__ */ N(Ft({
		state: "off",
		error: ""
	}));
	async function c() {
		try {
			P(n, await t.api.get("api/attention"), !0), P(r, "");
		} catch (e) {
			P(r, String(e.message || e), !0);
		}
		try {
			let e = await t.api.get("api/work?state=running&limit=6");
			P(i, await Promise.all((e.runs ?? []).slice(0, 6).map(async (e) => {
				try {
					let n = await t.api.get("api/ticket/" + encodeURIComponent(e.id)), r = n.comments ?? [];
					return {
						run: e,
						steps: n.run?.steps ?? [],
						comment: r.length ? r[r.length - 1] : null
					};
				} catch {
					return {
						run: e,
						steps: [],
						comment: null
					};
				}
			})), !0);
		} catch {}
		P(a, !0);
	}
	tn(() => {
		c(), t.api.push.status().then((e) => P(s, e, !0));
		let e = 0, n = t.api.subscribe(() => {
			e ||= window.setTimeout(() => {
				e = 0, c();
			}, 1e3);
		}), r = t.api.params();
		if (r.item && !r.session) {
			let [e, ...t] = r.item.split(":");
			l({
				kind: e,
				id: t.join(":")
			});
		}
		return () => {
			n(), clearTimeout(e);
		};
	});
	async function l(e) {
		let n = e.kind + ":" + e.id;
		P(o, n);
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
			P(o, "");
		}
	}
	let u = {
		unsupported: "This browser has no Web Push (on iOS, add Desk to the home screen first).",
		unavailable: "Desk runs without --push.",
		denied: "Notifications are blocked in the browser settings.",
		off: "Get a notification for each new attention item.",
		on: "Notifications are on for this browser."
	};
	async function d() {
		P(s, Y(s).state === "on" ? await t.api.push.unsubscribe() : await t.api.push.subscribe(), !0);
	}
	async function f() {
		P(s, await t.api.push.test(), !0), Y(s).error || t.api.toast("Test notification sent");
	}
	function p(e) {
		if (!e || e.startsWith("0001")) return "";
		let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3);
		return t < 60 ? t + "s ago" : t < 3600 ? Math.round(t / 60) + "m ago" : t < 86400 ? Math.round(t / 3600) + "h ago" : Math.round(t / 86400) + "d ago";
	}
	let m = (e, t) => typeof e?.[t] == "string" ? JSON.parse(e[t] || "{}") : e?.[t] || {};
	function h(e) {
		return e.map((e) => ({
			s: e,
			r: m(e, "meta").report
		})).filter((e) => e.r && (e.r.note || e.r.state)).sort((e, t) => String(t.r.at).localeCompare(String(e.r.at)));
	}
	let g = (e) => e === "failed" || e === "blocked" ? "bad" : e === "waiting" || e === "dropped" ? "warn" : e === "ready" ? "ok" : "accent", _ = (e) => !!e && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(e);
	var v = Xr(), y = Ut(v), b = R(I(y), 4), x = I(b), S = (e) => {
		var t = Mr(), n = Ut(t), r = (e) => {
			var t = jr();
			Un("click", t, f), Z(e, t);
		};
		$(n, (e) => {
			Y(s).state === "on" && e(r);
		});
		var i = R(n, 2);
		let a;
		var o = L(i, !0);
		z(() => {
			Dr(i, "title", u[Y(s).state]), a = Cr(i, 1, "", null, a, { on: Y(s).state === "on" }), Q(o, Y(s).state === "on" ? "Notifications on" : "Notify me");
		}), Un("click", i, d), Z(e, t);
	}, ee = (e) => {
		var t = Nr(), n = L(t, !0);
		z(() => {
			Dr(t, "title", u[Y(s).state]), Q(n, Y(s).state === "unavailable" ? "" : "no notifications");
		}), Z(e, t);
	};
	$(x, (e) => {
		Y(s).state === "on" || Y(s).state === "off" ? e(S) : e(ee, -1);
	}), O(b), O(y);
	var C = R(y, 2), w = I(C), te = (e) => {
		var t = Pr(), n = L(t, !0);
		z(() => Q(n, Y(s).error)), Z(e, t);
	};
	$(w, (e) => {
		Y(s).error && e(te);
	});
	var ne = R(w, 2), re = (e) => {
		var t = Pr(), n = L(t, !0);
		z(() => Q(n, Y(r))), Z(e, t);
	}, ie = (e) => {
		Z(e, Fr());
	}, ae = (e) => {
		Z(e, Ir());
	}, oe = (e) => {
		var t = Rr(), r = I(t), i = R(I(r));
		hr(i, 21, () => Y(n), (e) => e.kind + e.id, (e, t) => {
			var n = Lr();
			let r;
			var i = I(n), a = L(I(i), !0);
			O(i);
			var s = R(i), c = I(s, !0), u = R(c), d = I(u, !0), f = R(d), m = (e) => {
				var n = Qn();
				z(() => Q(n, `· ${Y(t).ref ?? ""}`)), Z(e, n);
			};
			$(f, (e) => {
				Y(t).ref && e(m);
			}), O(u), O(s);
			var h = R(s), g = L(h, !0), _ = L(R(h), !0);
			O(n), z((e) => {
				r = Cr(n, 1, "click svelte-t4vask", null, r, { busy: Y(o) === Y(t).kind + ":" + Y(t).id }), Dr(n, "data-item", Y(t).kind + ":" + Y(t).id), Q(a, Y(t).kind), Q(c, Y(t).title || Y(t).id), Q(d, Y(t).id), Q(g, Y(t).why), Q(_, e);
			}, [() => p(Y(t).since)]), Un("click", n, () => l(Y(t))), Z(e, n);
		}), O(i), O(r), O(t), Z(e, t);
	};
	$(ne, (e) => {
		Y(r) ? e(re) : Y(n) ? Y(n).length === 0 ? e(ae, 2) : e(oe, -1) : e(ie, 1);
	});
	var se = R(ne, 4), ce = (e) => {
		var n = qr();
		hr(n, 21, () => Y(i), ({ run: e, steps: t, comment: n }) => e.id, (e, n) => {
			let r = () => Y(n).run, i = () => Y(n).steps, a = () => Y(n).comment, o = /* @__PURE__ */ tt(() => h(i())), s = /* @__PURE__ */ tt(() => i().find((e) => e.status === "in_progress"));
			var c = Kr(), l = I(c), u = R(I(l), 2), d = L(u, !0), f = R(u, 2), v = L(f, !0), y = L(R(f, 2), !0);
			O(l);
			var b = R(l, 2), x = I(b);
			hr(x, 17, () => Y(o).slice(0, 3), ({ s: e, r: t }) => e.id, (e, n) => {
				let r = () => Y(n).s, i = () => Y(n).r;
				var a = Vr(), o = I(a), s = L(o, !0), c = R(o, 2), l = L(c, !0), u = R(c, 2), d = (e) => {
					var n = zr(), r = L(n, !0);
					z(() => {
						Dr(n, "href", "?session=" + i().by), Q(r, i().by);
					}), Un("click", n, (e) => {
						e.preventDefault(), t.api.openSession(i().by);
					}), Z(e, n);
				}, f = /* @__PURE__ */ tt(() => _(i().by));
				$(u, (e) => {
					Y(f) && e(d);
				});
				var h = R(u, 2), v = (e) => {
					var t = Br(), n = L(t, !0);
					z(() => Q(n, i().note)), Z(e, t);
				};
				$(h, (e) => {
					i().note && e(v);
				});
				var y = L(R(h, 2), !0);
				O(a), z((e, t, n) => {
					Cr(o, 1, `pill ${e ?? ""}`, "svelte-t4vask"), Q(s, i().state), Q(l, t), Q(y, n);
				}, [
					() => g(i().state),
					() => m(r(), "fields").step || r().title,
					() => p(i().at)
				]), Z(e, a);
			}, (e) => {
				var t = $n(), n = Ut(t), r = (e) => {
					var t = Hr(), n = I(t), r = L(n, !0), i = L(R(n), !0);
					be(), O(t), z((e) => {
						Q(r, Y(s).status), Q(i, e);
					}, [() => m(Y(s), "fields").step || Y(s).title]), Z(e, t);
				}, i = (e) => {
					Z(e, Ur());
				};
				$(n, (e) => {
					Y(s) ? e(r) : e(i, -1);
				}), Z(e, t);
			});
			var S = R(x, 2), ee = (e) => {
				var n = Gr(), r = R(I(n)), i = I(r), o = (e) => {
					var n = Wr(), r = L(n, !0);
					z(() => {
						Dr(n, "href", "?session=" + a().author), Q(r, a().author);
					}), Un("click", n, (e) => {
						e.preventDefault(), t.api.openSession(a().author);
					}), Z(e, n);
				}, s = /* @__PURE__ */ tt(() => _(a().author)), c = (e) => {
					var t = Qn();
					z(() => Q(t, a().author)), Z(e, t);
				};
				$(i, (e) => {
					Y(s) ? e(o) : e(c, -1);
				}), O(r);
				var l = R(r), u = L(l, !0), d = L(R(l), !0);
				O(n), z((e) => {
					Q(u, a().body), Q(d, e);
				}, [() => p(a().created_at)]), Z(e, n);
			};
			$(S, (e) => {
				a() && e(ee);
			}), O(b), O(c), z((e) => {
				Q(d, r().title), Q(v, r().id), Q(y, e);
			}, [() => p(r().updated_at)]), Un("click", l, () => t.api.openTicket(r().id)), Z(e, c);
		}), O(n), Z(e, n);
	}, le = (e) => {
		Z(e, Jr());
	}, ue = (e) => {
		Z(e, Yr());
	};
	$(se, (e) => {
		Y(i).length ? e(ce) : Y(a) ? e(ue, -1) : e(le, 1);
	}), O(C), Z(e, v), Re();
}
Wn(["click"]);
// plugins/attention/index.svelte.js
var Qr = new URL("./style.css", import.meta.url).href, $r = class extends HTMLElement {
	#e = /* @__PURE__ */ N(Ft({
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
		if (!document.querySelector(`link[href="${Qr}"]`)) {
			let e = document.createElement("link");
			e.rel = "stylesheet", e.href = Qr, document.head.append(e);
		}
		this.style.display = "contents", this.app = or(Zr, {
			target: this,
			props: this.props
		});
	}
	disconnectedCallback() {
		this.app && ur(this.app), this.app = null;
	}
};
customElements.get("desk-attention-view") || customElements.define("desk-attention-view", $r);
