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
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, ee = 1 << 25, C = 65536, te = 1 << 19, ne = 1 << 20, re = 1 << 25, ie = 1 << 21, ae = 1 << 22, oe = 1 << 23, se = Symbol("$state"), ce = Symbol("component"), le = Symbol("legacy props"), ue = Symbol(""), de = Symbol("attributes"), fe = Symbol("class"), pe = Symbol("style"), me = Symbol("text"), he = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), ge = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function _e() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function ve(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function ye() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var w = !1;
function be(e) {
	w = e;
}
var T;
function E(t) {
	if (t === null) throw ve(), e;
	return T = t;
}
function xe() {
	return E(/* @__PURE__ */ Kt(T));
}
function D(t) {
	if (w) {
		if (/* @__PURE__ */ Kt(T) !== null) throw ve(), e;
		T = t;
	}
}
function Se(e = 1) {
	if (w) {
		for (var t = e, n = T; t--;) n = /* @__PURE__ */ Kt(n);
		T = n;
	}
}
function Ce(e = !0) {
	for (var t = 0, n = T;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Kt(n);
		e && n.remove(), n = i;
	}
}
function we(t) {
	if (!t || t.nodeType !== 8) throw ve(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Te(e) {
	return e === this.v;
}
function Ee(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function De(e) {
	return !Ee(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function Oe() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function ke(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function Ae(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function je() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Me(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Ne() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Pe(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function Fe() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Ie() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Le() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Re() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var O = null;
function ze(e) {
	O = e;
}
function Be(e, t = !1, n) {
	O = {
		p: O,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: W,
		l: null
	};
}
function Ve(e) {
	var t = O, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) sn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, O = t.p, He(e);
}
function He(e = {}) {
	return s(e, ce, { value: !0 }), e;
}
function Ue() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var We = [];
function Ge() {
	var e = We;
	We = [], h(e);
}
function Ke(e) {
	if (We.length === 0 && !pt) {
		var t = We;
		queueMicrotask(() => {
			t === We && Ge();
		});
	}
	We.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var qe = ~(v | y | _);
function k(e, t) {
	e.f = e.f & qe | t;
}
function Je(e) {
	e.f & 512 || e.deps === null ? k(e, _) : k(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Ye(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), k(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Xe(e) {
	var t = V, n = W;
	U(null), Tn(null);
	try {
		return e();
	} finally {
		U(t), Tn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Ze(e, t, n, r) {
	let i = Ue() ? tt : it;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = W, c = Qe(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				$t(e, s);
			}
			$e();
		}
	}
	var d = et();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ rt(e))).then(u).catch((e) => $t(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), $e());
	}) : f();
}
function Qe() {
	var e = W, t = V, n = O, r = j;
	return function(i = !0) {
		Tn(e), U(t), ze(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function $e(e = !0) {
	Tn(null), U(null), ze(null), e && j?.deactivate();
}
function et() {
	var e = W, t = e.b, n = j, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function tt(e) {
	var n = 2 | v;
	return W !== null && (W.f |= te), {
		ctx: O,
		deps: null,
		effects: null,
		equals: Te,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: W,
		ac: null
	};
}
var nt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function rt(e, n, r) {
	let i = W;
	i === null && Oe();
	var a = void 0, o = Ot(t), s = !V, c = /* @__PURE__ */ new Set();
	return ln(() => {
		var t = W, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== he && n.reject(e);
			}).finally($e);
		} catch (e) {
			n.reject(e), $e();
		}
		var r = j;
		if (s) {
			if (t.f & 32768) var l = et();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(nt);
			else for (let e of c.values()) e.reject(nt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== nt && (r.activate(), t ? (o.f |= oe, Pt(o, t)) : (o.f & 8388608 && (o.f ^= oe), Pt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), an(() => {
		for (let e of c) e.reject(nt);
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
function A(e) {
	let t = /* @__PURE__ */ tt(e);
	return Dn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function it(e) {
	let t = /* @__PURE__ */ tt(e);
	return t.equals = De, t;
}
function at(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) B(t[n]);
	}
}
function ot(e) {
	var n, r = W, i = e.parent;
	if (!Cn && i !== null && e.v !== t && i.f & 24576) return _e(), e.v;
	Tn(i);
	try {
		at(e), n = In(e);
	} finally {
		Tn(r);
	}
	return n;
}
function st(e) {
	var t = ot(e);
	!e.equals(t) && (e.wv = Nn(), (!j?.is_fork || e.deps === null) && (j === null ? e.v = t : (j.capture(e, t, !0), dt?.capture(e, t, !0)), e.deps === null)) ? k(e, _) : Cn || (M === null ? Je(e) : (rn() || j?.is_fork) && M.set(e, t));
}
function ct(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Xe(() => {
		t.ac.abort(he), t.ac = null;
	}), t.fn !== null && (t.teardown = m), zn(t, 0), pn(t));
}
function lt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Bn(t);
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
			for (var r of n.d) k(r, v), t(r);
			for (r of n.m) k(r, y), t(r);
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
		for (let [e, t] of this.#s) k(e, t), this.schedule(e);
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
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (N.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : Pn(r) && (i & 16 && this.add_dirty_reaction(r, y), Bn(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), k(i, v), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), j = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Ye(e[t], this.#s);
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
			_t = 0, ft = null, ht = null, gt = null, mt = !1, j = null, M = null, N.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(nt);
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
		this.#u || (this.#u = !0, Ke(() => {
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
			!mt && Ke(() => {
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
		Ne();
	} catch (e) {
		$t(e, ft);
	}
}
var xt = null;
function St(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Pn(r) && (xt = /* @__PURE__ */ new Set(), Bn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && gn(r), xt?.size > 0)) {
				N.clear();
				for (let e of xt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) xt.has(n) && (xt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Bn(n);
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
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), k(e, _);
		for (var n = e.first; n !== null;) wt(n, t), n = n.next;
	}
}
function Tt(e) {
	k(e, _);
	for (var t = e.first; t !== null;) Tt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Et = /* @__PURE__ */ new Set(), N = /* @__PURE__ */ new Map(), Dt = !1;
function Ot(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Te,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function kt(e, t) {
	let n = Ot(e, t);
	return Dn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function At(e, t = !1, n = !0) {
	let r = Ot(e);
	return t || (r.equals = De), r;
}
function jt(e, t, n = !1) {
	return V !== null && (!H || V.f & 131072) && Ue() && V.f & 4325394 && (En === null || !En.has(e)) && Le(), Pt(e, n ? zt(t) : t, gt);
}
var Mt = null, Nt = 0;
function Pt(e, t, n = null) {
	if (!e.equals(t)) {
		Cn ? N.set(e, t) : N.has(e) || N.set(e, e.v);
		var r = yt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && ot(t), M === null && Je(t);
		}
		e.wv = Nn(), Mt = null, Nt = 0, Rt(e, v, n), Mt = null, Ue() && W !== null && W.f & 1024 && !(W.f & 96) && (q === null ? On([e]) : q.push(e)), !r.is_fork && Et.size > 0 && !Dt && Ft();
	}
	return t;
}
function Ft() {
	Dt = !1;
	for (let e of Et) {
		e.f & 1024 && k(e, y);
		let t;
		try {
			t = Pn(e);
		} catch {
			t = !0;
		}
		t && Bn(e);
	}
	Et.clear();
}
function It(e, t = 1) {
	var n = J(e), r = t === 1 ? n++ : n--;
	return jt(e, n), r;
}
function Lt(e) {
	jt(e, e.v + 1);
}
function Rt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ue(), a = r.length;
		if (Nt += a, Nt > 1e5 && Mt === null && (Mt = /* @__PURE__ */ new Set()), Mt !== null) {
			if (Mt.has(e)) return;
			Mt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== W) {
				var l = (c & v) === 0;
				if (l && k(s, t), c & 131072) Et.add(s);
				else if (c & 2) {
					var u = s;
					M?.delete(u), Rt(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && xt !== null && xt.add(d), n === null ? Ct(d) : n.push(d);
				}
			}
		}
	}
}
function zt(e) {
	if (typeof e != "object" || !e || se in e || ce in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ kt(0), s = null, l = jn, p = (e) => {
		if (jn === l) return e();
		var t = V, n = jn;
		U(null), Mn(l);
		var r = e();
		return U(t), Mn(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ kt(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Fe();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ kt(n.value, s);
				return i.set(t, e), e;
			}) : jt(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ kt(t, s));
					i.set(n, e), Lt(o);
				}
			} else jt(r, t), Lt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === se) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ kt(zt(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = J(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = J(a);
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
			if (n === se) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (W !== null || V !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ kt(a ? zt(e[n]) : t, s)), i.set(n, r)), J(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ kt(t, s)), i.set(f + "", m)) : jt(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ kt(void 0, s)), jt(u, zt(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => zt(r));
				jt(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && jt(_, v + 1);
				}
				Lt(o);
			}
			return !0;
		},
		ownKeys(e) {
			J(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = i.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [r, a] of i) a.v !== t && !(r in e) && n.push(r);
			return n;
		},
		setPrototypeOf() {
			Ie();
		}
	});
}
var Bt, Vt, Ht, Ut;
function Wt() {
	if (Bt === void 0) {
		Bt = window, Vt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Ht = c(t, "firstChild").get, Ut = c(t, "nextSibling").get, p(e) && (e[fe] = void 0, e[de] = null, e[pe] = void 0, e.__e = void 0), p(n) && (n[me] = void 0);
	}
}
function P(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Gt(e) {
	return Ht.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Kt(e) {
	return Ut.call(e);
}
function F(e, t) {
	if (!w) return /* @__PURE__ */ Gt(e);
	var n = /* @__PURE__ */ Gt(T);
	if (n === null) n = T.appendChild(P());
	else if (t && n.nodeType !== 3) {
		var r = P();
		return n?.before(r), E(r), r;
	}
	return t && Zt(n), E(n), n;
}
function qt(e, t = !1) {
	if (!w) {
		var n = /* @__PURE__ */ Gt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Kt(n) : n;
	}
	if (t) {
		if (T?.nodeType !== 3) {
			var r = P();
			return T?.before(r), E(r), r;
		}
		Zt(T);
	}
	return T;
}
function I(e, t = !1) {
	if (!w) return /* @__PURE__ */ Gt(e);
	var n = F(e, t);
	return D(e), n;
}
function L(e, t = 1, n = !1) {
	let r = w ? T : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Kt(r);
	if (!w) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = P();
			return r === null ? i?.after(a) : r.before(a), E(a), a;
		}
		Zt(r);
	}
	return E(r), r;
}
function Jt(e) {
	e.textContent = "";
}
function Yt() {
	return !1;
}
function Xt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function Zt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Qt(e) {
	var t = W;
	if (t === null) return V.f |= oe, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	$t(e, t);
}
function $t(t, n) {
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
function en(e) {
	W === null && (V === null && Me(e), je()), Cn && Ae(e);
}
function tn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function nn(e, t) {
	var n = W;
	n !== null && n.f & 8192 && (e |= b);
	var r = {
		ctx: O,
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
			Bn(r);
		} catch (e) {
			throw B(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= C));
	}
	if (i !== null && (i.parent = n, n !== null && tn(i, n), V !== null && V.f & 2 && !(e & 64))) {
		var a = V;
		(a.effects ??= []).push(i);
	}
	return r;
}
function rn() {
	return V !== null && !H;
}
function an(e) {
	let t = nn(8, null);
	return k(t, _), t.teardown = e, t;
}
function on(e) {
	en("$effect");
	var t = W.f;
	if (!V && t & 32 && O !== null && !O.i) {
		var n = O;
		(n.e ??= []).push(e);
	} else return sn(e);
}
function sn(e) {
	return nn(4 | ne, e);
}
function cn(e) {
	yt.ensure();
	let t = nn(64 | te, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? _n(t, () => {
			B(t), n(void 0);
		}) : (B(t), n(void 0));
	});
}
function ln(e) {
	return nn(ae | te, e);
}
function un(e, t = 0) {
	return nn(8 | t, e);
}
function R(e, t = [], n = [], r = []) {
	Ze(r, t, n, (t) => {
		nn(8, () => {
			e(...t.map(J));
		});
	});
}
function dn(e, t = 0) {
	return nn(16 | t, e);
}
function z(e) {
	return nn(32 | te, e);
}
function fn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Cn, r = V;
		wn(!0), U(null);
		try {
			t.call(null);
		} catch (t) {
			$t(t, e.parent);
		} finally {
			wn(n), U(r);
		}
	}
}
function pn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Xe(() => {
			e.abort(he);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : B(n, t), n = r;
	}
}
function mn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || B(t), t = n;
	}
}
function B(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (hn(e.nodes.start, e.nodes.end), n = !0), e.f |= ee, pn(e, t && !n), zn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	fn(e), e.f ^= ee, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && gn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function hn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Kt(e);
		e.remove(), e = n;
	}
}
function gn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function _n(e, t, n = !0) {
	var r = [];
	e.f |= 256, vn(e, r, !0);
	var i = () => {
		n && B(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function vn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= b;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				vn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function yn(e) {
	e.f &= -257, bn(e, !0);
}
function bn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= b, e.f & 1024 || (k(e, v), yt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			bn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function xn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Kt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Sn = null, Cn = !1;
function wn(e) {
	Cn = e;
}
var V = null, H = !1;
function U(e) {
	V = e;
}
var W = null;
function Tn(e) {
	W = e;
}
var En = null;
function Dn(e) {
	V !== null && (V.f & 2097152 || V.f & 2) && (En ??= /* @__PURE__ */ new Set()).add(e);
}
var G = null, K = 0, q = null;
function On(e) {
	q = e;
}
var kn = 1, An = 0, jn = An;
function Mn(e) {
	jn = e;
}
function Nn() {
	return ++kn;
}
function Pn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Pn(a) && st(a), a.wv > e.wv) return !0;
		}
		t & 512 && M === null && k(e, _);
	}
	return !1;
}
function Fn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(En !== null && En.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Fn(a, t, !1) : t === a && (n ? k(a, v) : a.f & 1024 && k(a, y), Ct(a));
	}
}
function In(e) {
	var t = G, n = K, r = q, i = V, a = En, o = O, s = H, c = jn, l = e.f;
	G = null, K = 0, q = null, V = l & 96 ? null : e, En = null, ze(e.ctx), H = !1, jn = ++An, e.ac !== null && (Xe(() => {
		e.ac.abort(he);
	}), e.ac = null);
	try {
		e.f |= ie;
		var u = e.fn, d = u();
		e.f |= S;
		var f = Ln(e);
		if (Ue() && q !== null && !H && f !== null && !(e.f & 6146)) for (var p = 0; p < q.length; p++) Fn(q[p], e);
		if (i !== null && i !== e) {
			if (An++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = An;
			if (t !== null) for (let e of t) e.rv = An;
			q !== null && (r === null ? r = q : r.push(...q));
		}
		return e.f & 8388608 && (e.f ^= oe), d;
	} catch (t) {
		return Ln(e), Qt(t);
	} finally {
		e.f ^= ie, G = t, K = n, q = r, V = i, En = a, ze(o), H = s, jn = c;
	}
}
function Ln(e) {
	var t = e.deps, n = j?.is_fork;
	if (G !== null) {
		var r;
		if (n || zn(e, K), t !== null && K > 0) for (t.length = K + G.length, r = 0; r < G.length; r++) t[K + r] = G[r];
		else e.deps = t = G;
		if (rn() && e.f & 512) for (r = K; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && K < t.length && (zn(e, K), t.length = K);
	return t;
}
function Rn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (G === null || !a.call(G, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Je(c), c.ac !== null && Xe(() => {
			c.ac.abort(he), c.ac = null, k(c, v);
		}), ct(c), zn(c, 0);
	}
}
function zn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Rn(e, n[r]);
}
function Bn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		k(e, _);
		var n = W;
		W = e;
		try {
			t & 16777232 ? mn(e) : pn(e), fn(e);
			var r = In(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = kn;
		} finally {
			W = n;
		}
	}
}
function J(e) {
	var t = !!(e.f & 2);
	if (Sn?.add(e), V !== null && !H && !(W !== null && W.f & 16384) && (En === null || !En.has(e))) {
		var n = V.deps;
		if (V.f & 2097152) e.rv < An && (e.rv = An, G === null && n !== null && n[K] === e ? K++ : G === null ? G = [e] : G.push(e));
		else {
			V.deps ??= [], a.call(V.deps, e) || V.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [V] : a.call(r, V) || r.push(V);
		}
	}
	if (Cn && N.has(e)) return N.get(e);
	if (t) {
		var i = e;
		if (Cn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Hn(i)) && (o = ot(i)), N.set(i, o), o;
		}
		var s = !(i.f & 512) && !H && V !== null && !!(V.f & 512), c = (i.f & S) === 0;
		Pn(i) && (s && (i.f |= 512), st(i)), s && !c && (lt(i), Vn(i));
	}
	if (M?.has(e)) return M.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Vn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (lt(n), Vn(n));
	}
}
function Hn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (N.has(t) || t.f & 2 && Hn(t)) return !0;
	return !1;
}
function Un(e) {
	var t = H;
	try {
		return H = !0, e();
	} finally {
		H = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Wn = Symbol("events"), Gn = /* @__PURE__ */ new Set(), Kn = /* @__PURE__ */ new Set();
function qn(e, t, n) {
	(t[Wn] ??= {})[e] = n;
}
function Jn(e) {
	for (var t = 0; t < e.length; t++) Gn.add(e[t]);
	for (var n of Kn) n(e);
}
var Yn = null, Xn = !1;
function Zn(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	Yn = e, Xn || (Xn = !0, setTimeout(() => {
		Xn = !1, Yn = null;
	}));
	var o = 0, c = Yn === e && e[Wn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Wn] = t;
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
		var d = V, f = W;
		U(null), Tn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Wn]?.[r];
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
			e[Wn] = t, delete e.currentTarget, U(d), Tn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var Qn = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function $n(e) {
	return Qn?.createHTML(e) ?? e;
}
function er(e) {
	var t = Xt("template");
	return t.innerHTML = $n(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function tr(e, t) {
	var n = W;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function Y(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (w) return tr(T, null), T;
		i === void 0 && (i = er(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Gt(i)));
		var t = r || Vt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Gt(t), s = t.lastChild;
			tr(o, s);
		} else tr(t, t);
		return t;
	};
}
function nr() {
	if (w) return tr(T, null), T;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = P();
	return e.append(t, n), tr(t, n), e;
}
function X(e, t) {
	if (w) {
		var n = W;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = T), xe();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var rr = ["touchstart", "touchmove"];
function ir(e) {
	return rr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function ar(e) {
	let t = 0, n = Ot(0), r;
	return () => {
		rn() && (J(n), un(() => (t === 0 && (r = Un(() => e(() => Lt(n)))), t += 1, () => {
			Ke(() => {
				--t, t === 0 && (r?.(), r = void 0, Lt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var or = C | te;
function sr(e, t, n, r) {
	new cr(e, t, n, r);
}
var cr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = w ? T : null;
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
	#m = ar(() => (this.#p = Ot(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = W;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = W.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = dn(() => {
			if (w) {
				let e = we(this.#t);
				xe();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, or), w && (this.#e = T);
	}
	#h() {
		try {
			this.#a = z(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Ke(r), t && (this.#s = z(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ye() : (t = !0, n && Re(), this.#s !== null && _n(this.#s, () => {
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
					$t(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | ee)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = z(() => e(this.#e)), Ke(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = P(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return z(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							$t(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(j)) : this.#u === 0 && (this.#e.before(e), this.#c = null, _n(this.#o, () => {
					this.#o = null;
				}), this.#x(j));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = z(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				xn(this.#a, e);
				let t = this.#n.pending;
				this.#o = z(() => t(this.#e));
			} else this.#x(j);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Ye(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = W, n = V, r = O;
		Tn(this.#i), U(this.#i), ze(this.#i.ctx);
		try {
			return yt.ensure(), e();
		} finally {
			Tn(t), U(n), ze(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && _n(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ke(() => {
			this.#d = !1, this.#p && Pt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), J(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		j?.is_fork ? (this.#a && j.skip_effect(this.#a), this.#o && j.skip_effect(this.#o), this.#s && j.skip_effect(this.#s), j.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (B(this.#a), null), this.#o &&= (B(this.#o), null), this.#s &&= (B(this.#s), null), w && (E(this.#t), Se(), E(Ce()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return z(() => {
						var r = W;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return $t(e, this.#i.parent), null;
				}
			}));
		};
		Ke(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					$t(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => $t(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Z(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[me] ??= e.nodeValue) && (e[me] = n, e.nodeValue = `${n}`);
}
function lr(e, t) {
	return dr(e, t);
}
var ur = /* @__PURE__ */ new Map();
function dr(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Wt();
	var u = void 0, d = cn(() => {
		var c = r ?? n.appendChild(P());
		sr(c, { pending: () => {} }, (n) => {
			Be({});
			var r = O;
			if (s && (r.c = s), a && (i.$$events = a), w && tr(n, null), u = t(n, i) || He(), w && (W.nodes.end = T, T === null || T.nodeType !== 8 || T.data !== "]")) throw ve(), e;
			Ve();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = ir(r);
					for (let e of [n, document]) {
						var a = ur.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), ur.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Zn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Gn)), Kn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = ur.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Zn), t.delete(e), t.size === 0 && ur.delete(r)) : t.set(e, i);
			}
			Kn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return fr.set(u, d), u;
}
var fr = /* @__PURE__ */ new WeakMap();
function pr(e, t) {
	let n = fr.get(e);
	return n ? (fr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var mr = class {
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
			if (n) yn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (yn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (B(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						xn(r, t), t.append(P()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else B(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), _n(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (B(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = j, r = Yt();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = P();
				i.append(a), this.#n.set(e, {
					effect: z(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, z(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else w && (this.anchor = T), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function Q(e, t, n = !1) {
	var r;
	w && (r = T, xe());
	var i = new mr(e), a = n ? C : 0;
	function o(e, t) {
		if (w) {
			var n = we(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Ce();
				E(a), i.anchor = a, be(!1), i.ensure(e, t), be(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	dn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function hr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		_n(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					gr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Jt(d), d.append(u), e.items.clear();
		}
		gr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function gr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= re, xn(a, document.createDocumentFragment())) : B(t[i], n);
	}
}
var _r;
function vr(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = w ? E(/* @__PURE__ */ Gt(u)) : u.appendChild(P());
	}
	w && xe();
	var d = null, f = /* @__PURE__ */ it(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = J(f);
			_.fallback = d, br(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= re, Sr(d, null, c)) : yn(d) : _n(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: dn(() => {
			var e = J(f), r = e.length;
			let o = !1;
			w && we(c) === "[!" != (r === 0) && (c = Ce(), E(c), be(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = j, v = Yt(), y = 0; y < r; y += 1) {
				w && T.nodeType === 8 && T.data === "]" && (c = T, o = !0, be(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && Pt(S.v, b), S.i && Pt(S.i, y), v && _.unskip_effect(S.e)) : (S = xr(l, m ? c : _r ??= P(), b, x, y, a, t, n), m || (S.e.f |= re), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = z(() => s(c)) : (d = z(() => s(_r ??= P())), d.f |= re)), r > u.size && ke("", "", ""), w && r > 0 && E(Ce()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && be(!0), J(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, w && (c = T);
}
function yr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function br(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = yr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (yn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= re, _ === l) Sr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), Cr(e, d, _), Cr(e, _, y), Sr(_, y, n), d = _, p = [], m = [], l = yr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Sr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					Cr(e, S.prev, ee.next), Cr(e, d, S), Cr(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), Sr(_, l, n), Cr(e, _.prev, _.next), Cr(e, _, d === null ? e.effect.first : d.next), Cr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = yr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = yr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (gr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var C = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || C.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && C.push(l), l = yr(l.next);
		var te = C.length;
		if (te > 0) {
			var ne = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.measure();
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.fix();
			}
			hr(e, C, ne);
		}
	}
	a && Ke(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function xr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Ot(n) : /* @__PURE__ */ At(n, !1, !1) : null, l = o & 2 ? Ot(i) : null;
	return {
		v: c,
		i: l,
		e: z(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Sr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Kt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function Cr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var wr = [..." 	\n\r\f\xA0\v﻿"];
function Tr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || wr.includes(r[o - 1])) && (s === r.length || wr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Er(e, t, n, r, i, a) {
	var o = e[fe];
	if (w || o !== n || o === void 0) {
		var s = Tr(n, r, a);
		w && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[fe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Dr = Symbol("is custom element"), Or = Symbol("is html"), kr = ge ? "link" : "LINK";
function $(e, t, n, r) {
	var i = Ar(e);
	if (!(w && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === kr))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[ue] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Mr(e).has(t) ? e[t] = n : (!w || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Ar(e) {
	return e[de] ??= {
		[Dr]: e.nodeName.includes("-"),
		[Or]: e.namespaceURI === n
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
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, l = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ tt(r), J(u)) : (l && (l = !1, s = o ? Un(r) : r), s);
	let f;
	if (a) {
		var p = se in e || le in e;
		f = c(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Pr(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Pe(t), f(m)));
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
	var v = !1, y = (n & 1 ? tt : it)(() => (v = !1, g()));
	a && J(y);
	var b = W;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? J(y) : i && a ? zt(e) : e;
			return jt(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return Cn && v || b.f & 16384 ? y.v : J(y);
	});
}
// kit/plugin.svelte.js
function Ir(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href, r = null, i = !1, a = () => (r ??= new Promise((e) => {
		let t = document.querySelector(`link[href="${n}"]`) ?? document.createElement("link");
		t.sheet ? e() : (t.addEventListener("load", () => e(), { once: !0 }), t.addEventListener("error", () => e(), { once: !0 }), t.isConnected || (t.rel = "stylesheet", t.href = n, document.head.append(t)));
	}).then(() => {
		i = !0;
	}), r);
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ kt(zt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return J(this.#e);
		}
		set props(e) {
			jt(this.#e, e, !0);
		}
		app = null;
		set api(e) {
			this.props.api = e;
		}
		get api() {
			return this.props.api;
		}
		set context(e) {
			this.props.context = e ?? {};
		}
		get context() {
			return this.props.context;
		}
		set config(e) {
			this.props.config = e ?? {};
		}
		get config() {
			return this.props.config;
		}
		connectedCallback() {
			this.style.display = "contents";
			let e = a();
			this.app ||= lr(t, {
				target: this,
				props: this.props
			}), i || (this.style.visibility = "hidden", e.then(() => {
				this.style.visibility = "";
			}));
		}
		disconnectedCallback() {
			this.app && pr(this.app), this.app = null;
		}
	});
}
function Lr(e, t) {
	let n = zt({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			It(r);
		}
	}), r = /* @__PURE__ */ kt(0);
	on(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, It(r);
			}, 1e3);
		});
		return () => {
			i(), clearTimeout(n);
		};
	});
	let i = !1, a = !1, o = "", s = (e, t) => {
		i = !0, a = !1;
		let r = (r) => {
			i = !1, t === o && (Object.assign(n, r), a && s(e, t));
		};
		e.get(t).then((e) => r({
			data: e,
			error: "",
			loading: !1
		}), (e) => r({
			error: String(e?.message || e),
			loading: !1
		}));
	};
	return on(() => {
		let n = e(), c = t();
		J(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/format.js
function Rr(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
function zr(e) {
	return !e || !e.fields ? {} : typeof e.fields == "string" ? JSON.parse(e.fields) : e.fields;
}
function Br(e) {
	return !e || !e.meta ? {} : typeof e.meta == "string" ? JSON.parse(e.meta) : e.meta;
}
function Vr(e) {
	switch (e) {
		case "done":
		case "running":
		case "closed": return "ok";
		case "needs_human":
		case "in_progress":
		case "request_changes": return "warn";
		case "failed":
		case "cancelled":
		case "crashed":
		case "stuck": return "bad";
		default: return "";
	}
}
function Hr(e) {
	let t = Br(e), n = [], r = t.github || {}, i = r.repo || "elee1766/jobshop";
	r.issue && n.push({
		label: "issue #" + r.issue,
		href: `https://github.com/${i}/issues/${r.issue}`
	}), r.pr && n.push({
		label: "PR #" + r.pr,
		href: `https://github.com/${i}/pull/${r.pr}`
	});
	let a = t.linear || {}, o = a.issue || a.identifier, s = a.url || (a.workspace && o ? `https://linear.app/${encodeURIComponent(a.workspace)}/issue/${encodeURIComponent(o)}` : "");
	return s && Ur(s) && n.push({
		label: o || "Linear",
		href: s
	}), n;
}
function Ur(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
function Wr(e) {
	return !!e && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(e);
}
function Gr(e, t) {
	let n = t.title || "", r = t.target || "", i = t.intent || "";
	return e?.name === "jcode" && i ? {
		label: i,
		detail: r || n
	} : n ? {
		label: n,
		detail: r || i
	} : {
		label: r || i || "tool",
		detail: r && i ? i : ""
	};
}
// kit/Holder.svelte
var Kr = /* @__PURE__ */ Y("<span class=\"svelte-bwb1ob\"> </span>"), qr = /* @__PURE__ */ Y("<span class=\"name svelte-bwb1ob\" data-testid=\"holder-pool\"> </span>"), Jr = /* @__PURE__ */ Y("<span class=\"name svelte-bwb1ob\" data-testid=\"holder-name\"> </span>"), Yr = /* @__PURE__ */ Y("<span class=\"faint model svelte-bwb1ob\"> </span>"), Xr = /* @__PURE__ */ Y("<span class=\"tool muted svelte-bwb1ob\" data-testid=\"holder-tool\"> </span>"), Zr = /* @__PURE__ */ Y("<button type=\"button\" class=\"holder svelte-bwb1ob\" data-testid=\"holder\"><span></span> <!> <!> <!> <!></button>");
function Qr(e, t) {
	Be(t, !0);
	let n = Fr(t, "label", 3, ""), r = /* @__PURE__ */ A(() => t.h.display_name || t.h.id), i = /* @__PURE__ */ A(() => t.h.tool || t.h.tool_target ? Gr(t.h.agent, {
		title: t.h.tool,
		target: t.h.tool_target,
		intent: t.h.tool_intent
	}).label : "working"), a = /* @__PURE__ */ A(() => [
		J(r),
		[t.h.pool, t.h.model].filter(Boolean).join(" · "),
		t.h.busy ? J(i) : "",
		t.h.id
	].filter(Boolean).join("\n") + "\nClick to open read-only");
	var o = Zr(), s = F(o);
	let c;
	var l = L(s, 2), u = (e) => {
		var t = Kr(), r = I(t, !0);
		R(() => Z(r, n())), X(e, t);
	};
	Q(l, (e) => {
		n() && e(u);
	});
	var d = L(l, 2), f = (e) => {
		var n = qr(), r = I(n, !0);
		R(() => Z(r, t.h.pool)), X(e, n);
	}, p = (e) => {
		var t = Jr(), n = I(t, !0);
		R(() => Z(n, J(r))), X(e, t);
	};
	Q(d, (e) => {
		t.h.pool ? e(f) : e(p, -1);
	});
	var m = L(d, 2), h = (e) => {
		var n = Yr(), r = I(n, !0);
		R(() => Z(r, t.h.model)), X(e, n);
	};
	Q(m, (e) => {
		t.h.model && e(h);
	});
	var g = L(m, 2), _ = (e) => {
		var t = Xr(), n = I(t, !0);
		R(() => Z(n, J(i))), X(e, t);
	};
	Q(g, (e) => {
		t.h.busy && e(_);
	}), D(o), R(() => {
		$(o, "title", J(a)), $(o, "data-session", t.h.id), $(o, "data-name", J(r)), c = Er(s, 1, "dot svelte-bwb1ob", null, c, {
			busy: t.h.busy,
			dead: t.h.alive === !1
		});
	}), qn("click", o, (e) => {
		e.stopPropagation(), t.onopen(t.h.id);
	}), X(e, o), Ve();
}
Jn(["click"]);
// kit/Trace.svelte
var $r = /* @__PURE__ */ Y("<span class=\"faint\"> </span>"), ei = /* @__PURE__ */ Y("<span data-testid=\"trace-session\"><!></span>"), ti = /* @__PURE__ */ Y("<a target=\"_blank\" rel=\"noopener noreferrer\" class=\"faint svelte-vk0dlq\"> </a>"), ni = /* @__PURE__ */ Y("<div class=\"kids svelte-vk0dlq\"></div>"), ri = /* @__PURE__ */ Y("<div class=\"node svelte-vk0dlq\"><a class=\"mono svelte-vk0dlq\"> </a> <span> </span> <span class=\"t svelte-vk0dlq\"> </span> <!> <!> <!> <!></div>");
function ii(e, t) {
	Be(t, !0);
	var n = ri(), r = F(n), i = I(r, !0), a = L(r, 2), o = I(a, !0), s = L(a, 2), c = I(s, !0), l = L(s, 2), u = (e) => {
		var n = $r(), r = I(n);
		R(() => Z(r, `via ${t.node.trigger ?? ""}`)), X(e, n);
	};
	Q(l, (e) => {
		t.node.trigger && e(u);
	});
	var d = L(l, 2), f = (e) => {
		var n = ei(), r = F(n);
		{
			let e = /* @__PURE__ */ A(() => t.node.session ?? { id: t.node.holder });
			Qr(r, {
				get h() {
					return J(e);
				},
				get onopen() {
					return t.watch;
				},
				label: "View session"
			});
		}
		D(n), X(e, n);
	};
	Q(d, (e) => {
		t.node.holder && t.watch && e(f);
	});
	var p = L(d, 2);
	vr(p, 16, () => t.node.links ?? [], (e) => e, (e, t) => {
		var n = nr(), r = qt(n), i = (e) => {
			var n = ti(), r = I(n, !0);
			R((e) => {
				$(n, "href", t), Z(r, e);
			}, [() => t.replace(/^https:\/\/github\.com\//, "")]), X(e, n);
		}, a = /* @__PURE__ */ A(() => Ur(t));
		Q(r, (e) => {
			J(a) && e(i);
		}), X(e, n);
	});
	var m = L(p, 2), h = (e) => {
		var n = ni();
		vr(n, 21, () => t.node.children, (e) => e.id, (e, n) => {
			ii(e, {
				get node() {
					return J(n);
				},
				get go() {
					return t.go;
				},
				get watch() {
					return t.watch;
				}
			});
		}), D(n), X(e, n);
	};
	Q(m, (e) => {
		t.node.children?.length && e(h);
	}), D(n), R((e) => {
		$(r, "href", "?ticket=" + t.node.id), Z(i, t.node.id), Er(a, 1, `pill ${e ?? ""}`, "svelte-vk0dlq"), Z(o, t.node.status), Z(c, t.node.title);
	}, [() => Vr(t.node.status)]), qn("click", r, (e) => {
		e.preventDefault(), t.go({ ticket: t.node.id });
	}), X(e, n), Ve();
}
Jn(["click"]);
// plugins/ticket/Details.svelte
var ai = /* @__PURE__ */ Y("<a class=\"mono sl svelte-jqeuhg\"> </a>"), oi = /* @__PURE__ */ Y("<span class=\"pill\"> </span>"), si = /* @__PURE__ */ Y("<span class=\"pill accent\"> </span>"), ci = /* @__PURE__ */ Y("<a target=\"_blank\" rel=\"noopener noreferrer\"> </a>"), li = /* @__PURE__ */ Y("<span class=\"muted held svelte-jqeuhg\" data-testid=\"held-by\">held by <!> <button type=\"button\" class=\"ghost\" data-testid=\"message-session\">Message this session</button></span>"), ui = /* @__PURE__ */ Y("<span class=\"mono\"> </span>"), di = /* @__PURE__ */ Y("<span class=\"muted\">held by <!></span>"), fi = /* @__PURE__ */ Y("<div class=\"row svelte-jqeuhg\"><span class=\"muted\">parent</span> <a class=\"mono\"> </a></div>"), pi = /* @__PURE__ */ Y("<span> </span>"), mi = /* @__PURE__ */ Y("<div class=\"report svelte-jqeuhg\"><span class=\"pill accent\"> </span> <!> <span class=\"faint\"> </span></div>"), hi = /* @__PURE__ */ Y("<span class=\"pill bad\"> </span>"), gi = /* @__PURE__ */ Y("<span class=\"pill warn\" title=\"request_changes closes of this step\"> </span>"), _i = /* @__PURE__ */ Y("<span class=\"pill warn\"> </span>"), vi = /* @__PURE__ */ Y("<div class=\"muted note svelte-jqeuhg\"> <span class=\"faint\"> </span></div>"), yi = /* @__PURE__ */ Y("<li class=\"svelte-jqeuhg\"><div class=\"stepline svelte-jqeuhg\"><a class=\"svelte-jqeuhg\"> </a> <span> </span> <!> <!> <!> <!></div> <!></li>"), bi = /* @__PURE__ */ Y("<h3 class=\"svelte-jqeuhg\">Steps</h3> <ol class=\"steps svelte-jqeuhg\"></ol>", 1), xi = /* @__PURE__ */ Y("<h3 class=\"svelte-jqeuhg\">Body</h3> <pre class=\"body svelte-jqeuhg\"> </pre>", 1), Si = /* @__PURE__ */ Y("<h3 class=\"svelte-jqeuhg\">Trace</h3> <!>", 1), Ci = /* @__PURE__ */ Y("<a target=\"_blank\" rel=\"noopener noreferrer\" class=\"svelte-jqeuhg\">open</a>"), wi = /* @__PURE__ */ Y("<li class=\"svelte-jqeuhg\"><span class=\"mono\"> </span> <span class=\"faint\"> </span> <!></li>"), Ti = /* @__PURE__ */ Y("<h3 class=\"svelte-jqeuhg\">External activity</h3> <ul class=\"ext svelte-jqeuhg\" data-testid=\"drawer-activity\"></ul>", 1), Ei = /* @__PURE__ */ Y("<div class=\"comment svelte-jqeuhg\"><div class=\"faint\"><!> </div> <pre> </pre></div>"), Di = /* @__PURE__ */ Y("<h3 class=\"svelte-jqeuhg\">Comments</h3> <!>", 1), Oi = /* @__PURE__ */ Y("<div class=\"ticket\"><h2 class=\"svelte-jqeuhg\"> </h2> <div class=\"row svelte-jqeuhg\"><span class=\"pill\"> </span> <span> </span> <!> <!></div> <div class=\"row svelte-jqeuhg\"><!> <!> <span class=\"faint\"> </span></div> <!> <!> <!> <!> <!> <!> <!></div>");
function ki(e, t) {
	Be(t, !0);
	let n = (e, t = m) => {
		var n = ai(), r = I(n, !0);
		R(() => {
			$(n, "href", "?session=" + t()), Z(r, t());
		}), qn("click", n, (e) => {
			e.preventDefault(), e.stopPropagation(), s(t());
		}), X(e, n);
	}, r = Fr(t, "context", 19, () => ({})), i = (e) => t.api.go(e), a = /* @__PURE__ */ A(() => r().ticket), o = Lr(() => t.api, () => J(a) ? "api/activity?limit=20&ticket=" + encodeURIComponent(J(a).id) : null), s = (e) => i({
		session: e,
		ticket: ""
	}), c = (e) => t.api.openSession(e, { readonly: !0 }), l = (e) => r().sessions?.[e] ?? { id: e }, u = (e) => Br(e).report || null;
	var d = nr(), f = qt(d), p = (e) => {
		let t = /* @__PURE__ */ A(() => zr(J(a))), d = /* @__PURE__ */ A(() => u(J(a)));
		var f = Oi(), p = F(f), m = I(p, !0), h = L(p, 2), g = F(h), _ = I(g, !0), v = L(g, 2), y = I(v), b = L(v, 2), x = (e) => {
			var t = oi(), n = I(t, !0);
			R(() => Z(n, J(a).pool)), X(e, t);
		};
		Q(b, (e) => {
			J(a).pool && e(x);
		}), vr(L(b, 2), 16, () => J(a).labels ?? [], (e) => e, (e, t) => {
			var n = si(), r = I(n, !0);
			R(() => Z(r, t)), X(e, n);
		}), D(h);
		var S = L(h, 2), ee = F(S);
		vr(ee, 17, () => Hr(J(a)), (e) => e.href, (e, t) => {
			var n = ci(), r = I(n, !0);
			R(() => {
				$(n, "href", J(t).href), Z(r, J(t).label);
			}), X(e, n);
		});
		var C = L(ee, 2), te = (e) => {
			var t = nr(), r = qt(t), i = (e) => {
				var t = li(), n = L(F(t));
				{
					let e = /* @__PURE__ */ A(() => l(J(a).holder));
					Qr(n, {
						get h() {
							return J(e);
						},
						onopen: c,
						label: "View session"
					});
				}
				var r = L(n, 2);
				D(t), qn("click", r, () => s(J(a).holder)), X(e, t);
			}, o = /* @__PURE__ */ A(() => Wr(J(a).holder) && J(a).status === "in_progress"), u = (e) => {
				var t = di(), r = L(F(t)), i = (e) => {
					n(e, () => J(a).holder);
				}, o = /* @__PURE__ */ A(() => Wr(J(a).holder)), s = (e) => {
					var t = ui(), n = I(t, !0);
					R(() => Z(n, J(a).holder)), X(e, t);
				};
				Q(r, (e) => {
					J(o) ? e(i) : e(s, -1);
				}), D(t), X(e, t);
			};
			Q(r, (e) => {
				J(o) ? e(i) : e(u, -1);
			}), X(e, t);
		};
		Q(C, (e) => {
			J(a).holder && e(te);
		});
		var ne = I(L(C, 2));
		D(S);
		var re = L(S, 2), ie = (e) => {
			var t = fi(), n = L(F(t), 2), r = I(n, !0);
			D(t), R(() => {
				$(n, "href", "?ticket=" + J(a).parent_id), Z(r, J(a).parent_id);
			}), qn("click", n, (e) => {
				e.preventDefault(), i({ ticket: J(a).parent_id });
			}), X(e, t);
		};
		Q(re, (e) => {
			J(a).parent_id && e(ie);
		});
		var ae = L(re, 2), oe = (e) => {
			var t = mi(), n = F(t), r = I(n, !0), i = L(n, 2), a = (e) => {
				var t = pi(), n = I(t, !0);
				R(() => Z(n, J(d).note)), X(e, t);
			};
			Q(i, (e) => {
				J(d).note && e(a);
			});
			var o = I(L(i, 2), !0);
			D(t), R((e) => {
				Z(r, J(d).state), Z(o, e);
			}, [() => Rr(J(d).at)]), X(e, t);
		};
		Q(ae, (e) => {
			J(d) && e(oe);
		});
		var se = L(ae, 2), ce = (e) => {
			var t = bi(), n = L(qt(t), 2);
			vr(n, 21, () => r().run.steps, (e) => e.id, (e, t) => {
				let n = /* @__PURE__ */ A(() => zr(J(t))), r = /* @__PURE__ */ A(() => u(J(t)));
				var a = yi(), o = F(a), s = F(o), d = I(s, !0), f = L(s, 2), p = I(f), m = L(f, 2), h = (e) => {
					var t = hi(), r = I(t);
					R(() => Z(r, `${J(n).failures ?? ""} failures`)), X(e, t);
				};
				Q(m, (e) => {
					J(n).failures && e(h);
				});
				var g = L(m, 2), _ = (e) => {
					var t = gi(), r = I(t);
					R(() => Z(r, `review round ${J(n).review_rounds ?? ""}/${J(n).max_review_rounds ?? 3 ?? ""}`)), X(e, t);
				};
				Q(g, (e) => {
					J(n).review_rounds && e(_);
				});
				var v = L(g, 2), y = (e) => {
					var t = _i(), r = I(t);
					R(() => {
						$(t, "title", J(n).changes.findings), Z(r, `changes requested by ${J(n).changes.step ?? ""}`);
					}), X(e, t);
				};
				Q(v, (e) => {
					J(n).changes && e(y);
				});
				var b = L(v, 2), x = (e) => {
					{
						let n = /* @__PURE__ */ A(() => l(J(t).holder));
						Qr(e, {
							get h() {
								return J(n);
							},
							onopen: c
						});
					}
				};
				Q(b, (e) => {
					J(t).status === "in_progress" && J(t).holder && e(x);
				}), D(o);
				var S = L(o, 2), ee = (e) => {
					var t = vi(), n = F(t), i = I(L(n), !0);
					D(t), R((e) => {
						Z(n, `${J(r).state ?? ""}: ${J(r).note ?? ""} `), Z(i, e);
					}, [() => Rr(J(r).at)]), X(e, t);
				};
				Q(S, (e) => {
					J(r)?.note && e(ee);
				}), D(a), R((e) => {
					$(s, "href", "?ticket=" + J(t).id), Z(d, J(n).step || J(t).title), Er(f, 1, `pill ${e ?? ""}`, "svelte-jqeuhg"), Z(p, `${J(t).status ?? ""}${J(t).resolution ? " · " + J(t).resolution : ""}`);
				}, [() => Vr(J(t).status)]), qn("click", s, (e) => {
					e.preventDefault(), i({ ticket: J(t).id });
				}), X(e, a);
			}), D(n), X(e, t);
		};
		Q(se, (e) => {
			r().run && e(ce);
		});
		var le = L(se, 2), ue = (e) => {
			var t = xi(), n = I(L(qt(t), 2), !0);
			R(() => Z(n, J(a).body)), X(e, t);
		};
		Q(le, (e) => {
			J(a).body && e(ue);
		});
		var de = L(le, 2), fe = (e) => {
			var t = Si();
			ii(L(qt(t), 2), {
				get node() {
					return r().trace;
				},
				go: i,
				watch: c
			}), X(e, t);
		};
		Q(de, (e) => {
			r().trace && e(fe);
		});
		var pe = L(de, 2), me = (e) => {
			var t = Ti(), n = L(qt(t), 2);
			vr(n, 21, () => o.data.items, (e) => e.id, (e, t) => {
				var n = wi(), r = F(n), i = I(r, !0), a = L(r), o = L(a), s = I(o), c = L(o, 2), l = (e) => {
					var n = Ci();
					R(() => $(n, "href", J(t).url)), X(e, n);
				}, u = /* @__PURE__ */ A(() => J(t).url && Ur(J(t).url));
				Q(c, (e) => {
					J(u) && e(l);
				}), D(n), R((e) => {
					Z(i, J(t).actor), Z(a, ` ${J(t).title ?? ""} `), Z(s, `· ${e ?? ""}`);
				}, [() => Rr(J(t).at)]), X(e, n);
			}), D(n), X(e, t);
		};
		Q(pe, (e) => {
			o.data?.items?.length && e(me);
		});
		var he = L(pe, 2), ge = (e) => {
			var t = Di();
			vr(L(qt(t), 2), 17, () => r().comments, (e) => e.id, (e, t) => {
				var r = Ei(), i = F(r), a = F(i), o = (e) => {
					n(e, () => J(t).author);
				}, s = /* @__PURE__ */ A(() => Wr(J(t).author)), c = (e) => {
					var n = ui(), r = I(n, !0);
					R(() => Z(r, J(t).author)), X(e, n);
				};
				Q(a, (e) => {
					J(s) ? e(o) : e(c, -1);
				});
				var l = L(a);
				D(i);
				var u = I(L(i, 2), !0);
				D(r), R((e) => {
					Z(l, ` · ${e ?? ""}`), Z(u, J(t).body);
				}, [() => Rr(J(t).created_at)]), X(e, r);
			}), X(e, t);
		};
		Q(he, (e) => {
			r().comments?.length && e(ge);
		}), D(f), R((e, n) => {
			Z(m, J(a).title), Z(_, J(a).kind), Er(v, 1, `pill ${e ?? ""}`, "svelte-jqeuhg"), Z(y, `${(J(t).state || J(a).status) ?? ""}${J(a).resolution ? " · " + J(a).resolution : ""}`), Z(ne, `updated ${n ?? ""}`);
		}, [() => Vr(J(t).state || J(a).status), () => Rr(J(a).updated_at)]), X(e, f);
	};
	Q(f, (e) => {
		J(a) && e(p);
	}), X(e, d), Ve();
}
// plugins/ticket/index.js
Jn(["click"]), Ir("desk-ticket-details", ki);
