// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = "http://www.w3.org/2000/svg", i = "http://www.w3.org/1998/Math/MathML", a = Array.isArray, o = Array.prototype.indexOf, s = Array.prototype.includes, c = Array.from, l = Object.defineProperty, u = Object.getOwnPropertyDescriptor, d = Object.getOwnPropertyDescriptors, f = Object.prototype, p = Array.prototype, m = Object.getPrototypeOf, h = Object.isExtensible, g = () => {};
function _(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function v() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
var y = 1024, b = 2048, x = 4096, S = 8192, C = 16384, ee = 32768, te = 1 << 25, ne = 65536, w = 1 << 19, re = 1 << 20, ie = 1 << 25, ae = 1 << 21, oe = 1 << 22, se = 1 << 23, ce = Symbol("$state"), le = Symbol("component"), ue = Symbol("legacy props"), de = Symbol(""), fe = Symbol("attributes"), pe = Symbol("class"), me = Symbol("style"), he = Symbol("text"), ge = Symbol("form reset"), T = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), _e = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function ve() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function ye(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function be() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var E = !1;
function xe(e) {
	E = e;
}
var D;
function Se(t) {
	if (t === null) throw ye(), e;
	return D = t;
}
function Ce() {
	return Se(/* @__PURE__ */ Zt(D));
}
function O(t) {
	if (E) {
		if (/* @__PURE__ */ Zt(D) !== null) throw ye(), e;
		D = t;
	}
}
function we(e = 1) {
	if (E) {
		for (var t = e, n = D; t--;) n = /* @__PURE__ */ Zt(n);
		D = n;
	}
}
function k(e = !0) {
	for (var t = 0, n = D;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Zt(n);
		e && n.remove(), n = i;
	}
}
function Te(t) {
	if (!t || t.nodeType !== 8) throw ye(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Ee(e) {
	return e === this.v;
}
function De(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Oe(e) {
	return !De(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function ke() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Ae(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function je(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Me() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Ne(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Pe() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Fe(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function Ie() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Le() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Re() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function ze() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var Be = null;
function Ve(e) {
	Be = e;
}
function He(e, t = !1, n) {
	Be = {
		p: Be,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: H,
		l: null
	};
}
function Ue(e) {
	var t = Be, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) fn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, Be = t.p, We(e);
}
function We(e = {}) {
	return l(e, le, { value: !0 }), e;
}
function Ge() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Ke = [];
function qe() {
	var e = Ke;
	Ke = [], _(e);
}
function Je(e) {
	if (Ke.length === 0 && !bt) {
		var t = Ke;
		queueMicrotask(() => {
			t === Ke && qe();
		});
	}
	Ke.push(e);
}
function Ye() {
	for (; Ke.length > 0;) qe();
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Xe = ~(b | x | y);
function A(e, t) {
	e.f = e.f & Xe | t;
}
function Ze(e) {
	e.f & 512 || e.deps === null ? A(e, y) : A(e, x);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Qe(e, t) {
	e.f & 2048 ? t.set(e, b) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, x), A(e, y);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
var $e = !1;
function et() {
	$e || ($e = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[ge]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function tt(e) {
	var t = V, n = H;
	Nn(null), Pn(null);
	try {
		return e();
	} finally {
		Nn(t), Pn(n);
	}
}
function nt(e, t, n, r = n) {
	e.addEventListener(t, () => tt(n));
	let i = e[ge];
	e[ge] = i ? () => {
		i(), r(!0);
	} : () => r(!0), et();
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function rt(e, t, n, r) {
	let i = Ge() ? st : dt;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = H, c = it(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				an(e, s);
			}
			at();
		}
	}
	var d = ot();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ lt(e))).then(u).catch((e) => an(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), at());
	}) : f();
}
function it() {
	var e = H, t = V, n = Be, r = M;
	return function(i = !0) {
		Pn(e), Nn(t), Ve(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function at(e = !0) {
	Pn(null), Nn(null), Ve(null), e && M?.deactivate();
}
function ot() {
	var e = H, t = e.b, n = M, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function st(e) {
	var n = 2 | b;
	return H !== null && (H.f |= w), {
		ctx: Be,
		deps: null,
		effects: null,
		equals: Ee,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: H,
		ac: null
	};
}
var ct = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function lt(e, n, r) {
	let i = H;
	i === null && ke();
	var a = void 0, o = Lt(t), s = !V, c = /* @__PURE__ */ new Set();
	return mn(() => {
		var t = H, n = v();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== T && n.reject(e);
			}).finally(at);
		} catch (e) {
			n.reject(e), at();
		}
		var r = M;
		if (s) {
			if (t.f & 32768) var l = ot();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(ct);
			else for (let e of c.values()) e.reject(ct);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== ct && (r.activate(), t ? (o.f |= se, Vt(o, t)) : (o.f & 8388608 && (o.f ^= se), Vt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), un(() => {
		for (let e of c) e.reject(ct);
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
function ut(e) {
	let t = /* @__PURE__ */ st(e);
	return In(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function dt(e) {
	let t = /* @__PURE__ */ st(e);
	return t.equals = Oe, t;
}
function ft(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) xn(t[n]);
	}
}
function pt(e) {
	var n, r = H, i = e.parent;
	if (!An && i !== null && e.v !== t && i.f & 24576) return ve(), e.v;
	Pn(i);
	try {
		ft(e), n = Jn(e);
	} finally {
		Pn(r);
	}
	return n;
}
function mt(e) {
	var t = pt(e);
	!e.equals(t) && (e.wv = Gn(), (!M?.is_fork || e.deps === null) && (M === null ? e.v = t : (M.capture(e, t, !0), _t?.capture(e, t, !0)), e.deps === null)) ? A(e, y) : An || (vt === null ? Ze(e) : (ln() || M?.is_fork) && vt.set(e, t));
}
function ht(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && tt(() => {
		t.ac.abort(T), t.ac = null;
	}), t.fn !== null && (t.teardown = g), Zn(t, 0), yn(t));
}
function j(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Qn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var gt = null, M = null, _t = null, vt = null, yt = null, bt = !1, xt = !1, St = null, Ct = null, wt = 0, Tt = 1, Et = class e {
	id = Tt++;
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
		gt === null ? gt = this : (gt.next = this, this.prev = gt), gt = this;
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
			for (var r of n.d) A(r, b), t(r);
			for (r of n.m) A(r, x), t(r);
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
					t.f ^= y;
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
		for (var t = St = [], n = [], r = Ct = []; this.#a.length > 0;) {
			wt++ > 1e3 && (this.#b(), Ot());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Nt(e), this.#d() || this.discard(), t;
			}
		}
		if (M = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (St = null, Ct = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Mt(e, t);
			r.length > 0 && M.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), _t = this, At(n), At(t), _t = null, this.#i?.resolve();
			var o = M;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Ft.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= y;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= y : i & 4 ? t.push(r) : Kn(r) && (i & 16 && this.add_dirty_reaction(r, x), Qn(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), A(i, b), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), M = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Qe(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), vt?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		M = this;
	}
	deactivate() {
		M = null, vt = null;
	}
	flush() {
		try {
			xt = !0, M = this, this.#m();
		} finally {
			wt = 0, yt = null, St = null, Ct = null, xt = !1, M = null, vt = null, Ft.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(ct);
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
		this.#u || (this.#u = !0, Je(() => {
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
		return (this.#i ??= v()).promise;
	}
	static ensure() {
		if (M === null) {
			let t = M = new e();
			!xt && !bt && Je(() => {
				t.started || t.flush();
			});
		}
		return M;
	}
	apply() {
		vt = null;
	}
	schedule(e) {
		yt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? gt = e : t.prev = e, this.linked = !1;
		}
	}
};
function Dt(e) {
	var t = bt, n = _t;
	_t = null, bt = !0;
	try {
		var r;
		for (e && (Dt(), r = e());;) {
			if (Ye(), M === null) return r;
			M.flush();
		}
	} finally {
		bt = t, _t = n;
	}
}
function Ot() {
	try {
		Pe();
	} catch (e) {
		an(e, yt);
	}
}
var kt = null;
function At(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Kn(r) && (kt = /* @__PURE__ */ new Set(), Qn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && Cn(r), kt?.size > 0)) {
				Ft.clear();
				for (let e of kt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) kt.has(n) && (kt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Qn(n);
					}
				}
				kt.clear();
			}
		}
		kt = null;
	}
}
function jt(e) {
	M.schedule(e);
}
function Mt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), A(e, y);
		for (var n = e.first; n !== null;) Mt(n, t), n = n.next;
	}
}
function Nt(e) {
	A(e, y);
	for (var t = e.first; t !== null;) Nt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Pt = /* @__PURE__ */ new Set(), Ft = /* @__PURE__ */ new Map(), It = !1;
function Lt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Ee,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function N(e, t) {
	let n = Lt(e, t);
	return In(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Rt(e, t = !1, n = !0) {
	let r = Lt(e);
	return t || (r.equals = Oe), r;
}
function P(e, t, n = !1) {
	return V !== null && (!Mn || V.f & 131072) && Ge() && V.f & 4325394 && (Fn === null || !Fn.has(e)) && Re(), Vt(e, n ? F(t) : t, Ct);
}
var zt = null, Bt = 0;
function Vt(e, t, n = null) {
	if (!e.equals(t)) {
		An ? Ft.set(e, t) : Ft.has(e) || Ft.set(e, e.v);
		var r = Et.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && pt(t), vt === null && Ze(t);
		}
		e.wv = Gn(), zt = null, Bt = 0, Wt(e, b, n), zt = null, Ge() && H !== null && H.f & 1024 && !(H.f & 96) && (zn === null ? Bn([e]) : zn.push(e)), !r.is_fork && Pt.size > 0 && !It && Ht();
	}
	return t;
}
function Ht() {
	It = !1;
	for (let e of Pt) {
		e.f & 1024 && A(e, x);
		let t;
		try {
			t = Kn(e);
		} catch {
			t = !0;
		}
		t && Qn(e);
	}
	Pt.clear();
}
function Ut(e) {
	P(e, e.v + 1);
}
function Wt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ge(), a = r.length;
		if (Bt += a, Bt > 1e5 && zt === null && (zt = /* @__PURE__ */ new Set()), zt !== null) {
			if (zt.has(e)) return;
			zt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== H) {
				var l = (c & b) === 0;
				if (l && A(s, t), c & 131072) Pt.add(s);
				else if (c & 2) {
					var u = s;
					vt?.delete(u), Wt(u, x, n);
				} else if (l) {
					var d = s;
					c & 16 && kt !== null && kt.add(d), n === null ? jt(d) : n.push(d);
				}
			}
		}
	}
}
function F(e) {
	if (typeof e != "object" || !e || ce in e || le in e) return e;
	let n = m(e);
	if (n !== f && n !== p) return e;
	var r = /* @__PURE__ */ new Map(), i = a(e), o = /* @__PURE__ */ N(0), s = null, c = Un, l = (e) => {
		if (Un === c) return e();
		var t = V, n = Un;
		Nn(null), Wn(c);
		var r = e();
		return Nn(t), Wn(n), r;
	};
	return i && r.set("length", /* @__PURE__ */ N(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Ie();
			var i = r.get(t);
			return i === void 0 ? l(() => {
				var e = /* @__PURE__ */ N(n.value, s);
				return r.set(t, e), e;
			}) : P(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = l(() => /* @__PURE__ */ N(t, s));
					r.set(n, e), Ut(o);
				}
			} else P(i, t), Ut(o);
			return !0;
		},
		get(n, i, a) {
			if (i === ce) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || u(n, i)?.writable) && (o = l(() => /* @__PURE__ */ N(F(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
				var d = U(o);
				return d === t ? void 0 : d;
			}
			return Reflect.get(n, i, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var i = Reflect.getOwnPropertyDescriptor(e, n), a = r.get(n);
			if (a !== void 0) {
				var o = U(a);
				if (o === t) return;
				if (i && "value" in i) i.value = o;
				else return {
					enumerable: !0,
					configurable: !0,
					value: o,
					writable: !0
				};
			}
			return i;
		},
		has(e, n) {
			if (n === ce) return !0;
			var i = r.get(n), a = i !== void 0 && i.v !== t || Reflect.has(e, n);
			return (i !== void 0 || (H !== null || V !== null) && (!a || u(e, n)?.writable)) && (i === void 0 && (i = l(() => /* @__PURE__ */ N(a ? F(e[n]) : t, s)), r.set(n, i)), U(i) === t) ? !1 : a;
		},
		set(e, n, a, c) {
			var d = r.get(n), f = n in e;
			if (i && n === "length") for (var p = a; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = l(() => /* @__PURE__ */ N(t, s)), r.set(p + "", m)) : P(m, t);
			}
			if (d === void 0) (!f || u(e, n)?.writable) && (d = l(() => /* @__PURE__ */ N(void 0, s)), P(d, F(a)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = l(() => F(a));
				P(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, a), !f) {
				if (i && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && P(_, v + 1);
				}
				Ut(o);
			}
			return !0;
		},
		ownKeys(e) {
			U(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = r.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [i, a] of r) a.v !== t && !(i in e) && n.push(i);
			return n;
		},
		setPrototypeOf() {
			Le();
		}
	});
}
var Gt, Kt, qt, Jt;
function Yt() {
	if (Gt === void 0) {
		Gt = window, Kt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		qt = u(t, "firstChild").get, Jt = u(t, "nextSibling").get, h(e) && (e[pe] = void 0, e[fe] = null, e[me] = void 0, e.__e = void 0), h(n) && (n[he] = void 0);
	}
}
function I(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Xt(e) {
	return qt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Zt(e) {
	return Jt.call(e);
}
function L(e, t) {
	if (!E) return /* @__PURE__ */ Xt(e);
	var n = /* @__PURE__ */ Xt(D);
	if (n === null) n = D.appendChild(I());
	else if (t && n.nodeType !== 3) {
		var r = I();
		return n?.before(r), Se(r), r;
	}
	return t && nn(n), Se(n), n;
}
function Qt(e, t = !1) {
	if (!E) {
		var n = /* @__PURE__ */ Xt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Zt(n) : n;
	}
	if (t) {
		if (D?.nodeType !== 3) {
			var r = I();
			return D?.before(r), Se(r), r;
		}
		nn(D);
	}
	return D;
}
function R(e, t = !1) {
	if (!E) return /* @__PURE__ */ Xt(e);
	var n = L(e, t);
	return O(e), n;
}
function z(e, t = 1, n = !1) {
	let r = E ? D : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Zt(r);
	if (!E) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = I();
			return r === null ? i?.after(a) : r.before(a), Se(a), a;
		}
		nn(r);
	}
	return Se(r), r;
}
function $t(e) {
	e.textContent = "";
}
function en() {
	return !1;
}
function tn(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function nn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function rn(e) {
	var t = H;
	if (t === null) return V.f |= se, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	an(e, t);
}
function an(t, n) {
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
function on(e) {
	H === null && (V === null && Ne(e), Me()), An && je(e);
}
function sn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function cn(e, t) {
	var n = H;
	n !== null && n.f & 8192 && (e |= S);
	var r = {
		ctx: Be,
		deps: null,
		nodes: null,
		f: e | b | 512,
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
	M?.register_created_effect(r);
	var i = r;
	if (e & 4) St === null ? Et.ensure().schedule(r) : St.push(r);
	else if (t !== null) {
		try {
			Qn(r);
		} catch (e) {
			throw xn(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ne));
	}
	if (i !== null && (i.parent = n, n !== null && sn(i, n), V !== null && V.f & 2 && !(e & 64))) {
		var a = V;
		(a.effects ??= []).push(i);
	}
	return r;
}
function ln() {
	return V !== null && !Mn;
}
function un(e) {
	let t = cn(8, null);
	return A(t, y), t.teardown = e, t;
}
function dn(e) {
	on("$effect");
	var t = H.f;
	if (!V && t & 32 && Be !== null && !Be.i) {
		var n = Be;
		(n.e ??= []).push(e);
	} else return fn(e);
}
function fn(e) {
	return cn(4 | re, e);
}
function pn(e) {
	Et.ensure();
	let t = cn(64 | w, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? wn(t, () => {
			xn(t), n(void 0);
		}) : (xn(t), n(void 0));
	});
}
function mn(e) {
	return cn(oe | w, e);
}
function hn(e, t = 0) {
	return cn(8 | t, e);
}
function B(e, t = [], n = [], r = []) {
	rt(r, t, n, (t) => {
		cn(8, () => {
			e(...t.map(U));
		});
	});
}
function gn(e, t = 0) {
	return cn(16 | t, e);
}
function _n(e) {
	return cn(32 | w, e);
}
function vn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = An, r = V;
		jn(!0), Nn(null);
		try {
			t.call(null);
		} catch (t) {
			an(t, e.parent);
		} finally {
			jn(n), Nn(r);
		}
	}
}
function yn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && tt(() => {
			e.abort(T);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : xn(n, t), n = r;
	}
}
function bn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || xn(t), t = n;
	}
}
function xn(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (Sn(e.nodes.start, e.nodes.end), n = !0), e.f |= te, yn(e, t && !n), Zn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	vn(e), e.f ^= te, e.f |= C;
	var i = e.parent;
	i !== null && i.first !== null && Cn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function Sn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Zt(e);
		e.remove(), e = n;
	}
}
function Cn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function wn(e, t, n = !0) {
	var r = [];
	e.f |= 256, Tn(e, r, !0);
	var i = () => {
		n && xn(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function Tn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= S;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				Tn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function En(e) {
	e.f &= -257, Dn(e, !0);
}
function Dn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= S, e.f & 1024 || (A(e, b), Et.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			Dn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function On(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Zt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var kn = null, An = !1;
function jn(e) {
	An = e;
}
var V = null, Mn = !1;
function Nn(e) {
	V = e;
}
var H = null;
function Pn(e) {
	H = e;
}
var Fn = null;
function In(e) {
	V !== null && (V.f & 2097152 || V.f & 2) && (Fn ??= /* @__PURE__ */ new Set()).add(e);
}
var Ln = null, Rn = 0, zn = null;
function Bn(e) {
	zn = e;
}
var Vn = 1, Hn = 0, Un = Hn;
function Wn(e) {
	Un = e;
}
function Gn() {
	return ++Vn;
}
function Kn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Kn(a) && mt(a), a.wv > e.wv) return !0;
		}
		t & 512 && vt === null && A(e, y);
	}
	return !1;
}
function qn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Fn !== null && Fn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? qn(a, t, !1) : t === a && (n ? A(a, b) : a.f & 1024 && A(a, x), jt(a));
	}
}
function Jn(e) {
	var t = Ln, n = Rn, r = zn, i = V, a = Fn, o = Be, s = Mn, c = Un, l = e.f;
	Ln = null, Rn = 0, zn = null, V = l & 96 ? null : e, Fn = null, Ve(e.ctx), Mn = !1, Un = ++Hn, e.ac !== null && (tt(() => {
		e.ac.abort(T);
	}), e.ac = null);
	try {
		e.f |= ae;
		var u = e.fn, d = u();
		e.f |= ee;
		var f = Yn(e);
		if (Ge() && zn !== null && !Mn && f !== null && !(e.f & 6146)) for (var p = 0; p < zn.length; p++) qn(zn[p], e);
		if (i !== null && i !== e) {
			if (Hn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Hn;
			if (t !== null) for (let e of t) e.rv = Hn;
			zn !== null && (r === null ? r = zn : r.push(...zn));
		}
		return e.f & 8388608 && (e.f ^= se), d;
	} catch (t) {
		return Yn(e), rn(t);
	} finally {
		e.f ^= ae, Ln = t, Rn = n, zn = r, V = i, Fn = a, Ve(o), Mn = s, Un = c;
	}
}
function Yn(e) {
	var t = e.deps, n = M?.is_fork;
	if (Ln !== null) {
		var r;
		if (n || Zn(e, Rn), t !== null && Rn > 0) for (t.length = Rn + Ln.length, r = 0; r < Ln.length; r++) t[Rn + r] = Ln[r];
		else e.deps = t = Ln;
		if (ln() && e.f & 512) for (r = Rn; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && Rn < t.length && (Zn(e, Rn), t.length = Rn);
	return t;
}
function Xn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var i = o.call(r, e);
		if (i !== -1) {
			var a = r.length - 1;
			a === 0 ? r = n.reactions = null : (r[i] = r[a], r.pop());
		}
	}
	if (r === null && n.f & 2 && (Ln === null || !s.call(Ln, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Ze(c), c.ac !== null && tt(() => {
			c.ac.abort(T), c.ac = null, A(c, b);
		}), ht(c), Zn(c, 0);
	}
}
function Zn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Xn(e, n[r]);
}
function Qn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		A(e, y);
		var n = H;
		H = e;
		try {
			t & 16777232 ? bn(e) : yn(e), vn(e);
			var r = Jn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Vn;
		} finally {
			H = n;
		}
	}
}
async function $n() {
	await Promise.resolve(), Dt();
}
function U(e) {
	var t = !!(e.f & 2);
	if (kn?.add(e), V !== null && !Mn && !(H !== null && H.f & 16384) && (Fn === null || !Fn.has(e))) {
		var n = V.deps;
		if (V.f & 2097152) e.rv < Hn && (e.rv = Hn, Ln === null && n !== null && n[Rn] === e ? Rn++ : Ln === null ? Ln = [e] : Ln.push(e));
		else {
			V.deps ??= [], s.call(V.deps, e) || V.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [V] : s.call(r, V) || r.push(V);
		}
	}
	if (An && Ft.has(e)) return Ft.get(e);
	if (t) {
		var i = e;
		if (An) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || tr(i)) && (a = pt(i)), Ft.set(i, a), a;
		}
		var o = !(i.f & 512) && !Mn && V !== null && !!(V.f & 512), c = (i.f & ee) === 0;
		Kn(i) && (o && (i.f |= 512), mt(i)), o && !c && (j(i), er(i));
	}
	if (vt?.has(e)) return vt.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function er(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : s.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (j(n), er(n));
	}
}
function tr(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Ft.has(t) || t.f & 2 && tr(t)) return !0;
	return !1;
}
function nr(e) {
	var t = Mn;
	try {
		return Mn = !0, e();
	} finally {
		Mn = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var rr = Symbol("events"), ir = /* @__PURE__ */ new Set(), ar = /* @__PURE__ */ new Set();
function or(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || fr.call(t, e), !e.cancelBubble) return tt(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, Je(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function sr(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = or(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && un(() => {
		o.__removed = !0, t.removeEventListener(e, o, a);
	});
}
function cr(e, t, n) {
	(t[rr] ??= {})[e] = n;
}
function lr(e) {
	for (var t = 0; t < e.length; t++) ir.add(e[t]);
	for (var n of ar) n(e);
}
var ur = null, dr = !1;
function fr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	ur = e, dr || (dr = !0, setTimeout(() => {
		dr = !1, ur = null;
	}));
	var o = 0, s = ur === e && e[rr];
	if (s) {
		var c = i.indexOf(s);
		if (c !== -1 && (t === document || t === window)) {
			e[rr] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		c <= u && (o = c);
	}
	if (a = i[o] || e.target, a !== t) {
		l(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = V, f = H;
		Nn(null), Pn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[rr]?.[r];
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
			e[rr] = t, delete e.currentTarget, Nn(d), Pn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var pr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function mr(e) {
	return pr?.createHTML(e) ?? e;
}
function hr(e) {
	var t = tn("template");
	return t.innerHTML = mr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function gr(e, t) {
	var n = H;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function W(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (E) return gr(D, null), D;
		i === void 0 && (i = hr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Xt(i)));
		var t = r || Kt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Xt(t), s = t.lastChild;
			gr(o, s);
		} else gr(t, t);
		return t;
	};
}
function _r(e = "") {
	if (!E) {
		var t = I(e + "");
		return gr(t, t), t;
	}
	var n = D;
	return n.nodeType === 3 ? nn(n) : (n.before(n = I()), Se(n)), gr(n, n), n;
}
function vr() {
	if (E) return gr(D, null), D;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = I();
	return e.append(t, n), gr(t, n), e;
}
function G(e, t) {
	if (E) {
		var n = H;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = D), Ce();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var yr = ["touchstart", "touchmove"];
function br(e) {
	return yr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function xr(e) {
	let t = 0, n = Lt(0), r;
	return () => {
		ln() && (U(n), hn(() => (t === 0 && (r = nr(() => e(() => Ut(n)))), t += 1, () => {
			Je(() => {
				--t, t === 0 && (r?.(), r = void 0, Ut(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var Sr = ne | w;
function Cr(e, t, n, r) {
	new wr(e, t, n, r);
}
var wr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = E ? D : null;
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
	#m = xr(() => (this.#p = Lt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = H;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = H.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = gn(() => {
			if (E) {
				let e = Te(this.#t);
				Ce();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, Sr), E && (this.#e = D);
	}
	#h() {
		try {
			this.#a = _n(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Je(r), t && (this.#s = _n(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? be() : (t = !0, n && ze(), this.#s !== null && wn(this.#s, () => {
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
					an(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (C | te)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = _n(() => e(this.#e)), Je(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = I(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return _n(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							an(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(M)) : this.#u === 0 && (this.#e.before(e), this.#c = null, wn(this.#o, () => {
					this.#o = null;
				}), this.#x(M));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = _n(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				On(this.#a, e);
				let t = this.#n.pending;
				this.#o = _n(() => t(this.#e));
			} else this.#x(M);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Qe(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = H, n = V, r = Be;
		Pn(this.#i), Nn(this.#i), Ve(this.#i.ctx);
		try {
			return Et.ensure(), e();
		} finally {
			Pn(t), Nn(n), Ve(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && wn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Je(() => {
			this.#d = !1, this.#p && Vt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), U(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		M?.is_fork ? (this.#a && M.skip_effect(this.#a), this.#o && M.skip_effect(this.#o), this.#s && M.skip_effect(this.#s), M.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (xn(this.#a), null), this.#o &&= (xn(this.#o), null), this.#s &&= (xn(this.#s), null), E && (Se(this.#t), we(), Se(k()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return _n(() => {
						var r = H;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return an(e, this.#i.parent), null;
				}
			}));
		};
		Je(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					an(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => an(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function K(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[he] ??= e.nodeValue) && (e[he] = n, e.nodeValue = `${n}`);
}
function Tr(e, t) {
	return Dr(e, t);
}
var Er = /* @__PURE__ */ new Map();
function Dr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: s = !0, transformError: l }) {
	Yt();
	var u = void 0, d = pn(() => {
		var s = r ?? n.appendChild(I());
		Cr(s, { pending: () => {} }, (n) => {
			He({});
			var r = Be;
			if (o && (r.c = o), a && (i.$$events = a), E && gr(n, null), u = t(n, i) || We(), E && (H.nodes.end = D, D === null || D.nodeType !== 8 || D.data !== "]")) throw ye(), e;
			Ue();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = br(r);
					for (let e of [n, document]) {
						var a = Er.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), Er.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, fr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(c(ir)), ar.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = Er.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, fr), t.delete(e), t.size === 0 && Er.delete(r)) : t.set(e, i);
			}
			ar.delete(f), s !== r && s.parentNode?.removeChild(s);
		};
	});
	return Or.set(u, d), u;
}
var Or = /* @__PURE__ */ new WeakMap();
function kr(e, t) {
	let n = Or.get(e);
	return n ? (Or.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Ar = class {
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
			if (n) En(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (En(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (xn(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						On(r, t), t.append(I()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else xn(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), wn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (xn(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = M, r = en();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = I();
				i.append(a), this.#n.set(e, {
					effect: _n(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, _n(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else E && (this.anchor = D), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function q(e, t, n = !1) {
	var r;
	E && (r = D, Ce());
	var i = new Ar(e), a = n ? ne : 0;
	function o(e, t) {
		if (E) {
			var n = Te(r);
			if (e !== parseInt(n.substring(1))) {
				var a = k();
				Se(a), i.anchor = a, xe(!1), i.ensure(e, t), xe(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	gn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function jr(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, s = 0; s < i; s++) {
		let n = t[s];
		wn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					Mr(e, c(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --o;
		}, !1);
	}
	if (o === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			$t(d), d.append(u), e.items.clear();
		}
		Mr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function Mr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= ie, On(a, document.createDocumentFragment())) : xn(t[i], n);
	}
}
var Nr;
function Pr(e, t, n, r, i, o = null) {
	var s = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		s = E ? Se(/* @__PURE__ */ Xt(u)) : u.appendChild(I());
	}
	E && Ce();
	var d = null, f = /* @__PURE__ */ dt(() => {
		var e = n();
		return a(e) ? e : e == null ? [] : c(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = U(f);
			_.fallback = d, Ir(_, n, s, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ie, Rr(d, null, s)) : En(d) : wn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: gn(() => {
			var e = U(f), a = e.length;
			let c = !1;
			E && Te(s) === "[!" != (a === 0) && (s = k(), Se(s), xe(!1), c = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = M, v = en(), y = 0; y < a; y += 1) {
				E && D.nodeType === 8 && D.data === "]" && (s = D, c = !0, xe(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Vt(S.v, b), S.i && Vt(S.i, y), v && _.unskip_effect(S.e)) : (S = Lr(l, m ? s : Nr ??= I(), b, x, y, i, t, n), m || (S.e.f |= ie), l.set(x, S)), u.add(x);
			}
			if (a === 0 && o && !d && (m ? d = _n(() => o(s)) : (d = _n(() => o(Nr ??= I())), d.f |= ie)), a > u.size && Ae("", "", ""), E && a > 0 && Se(k()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			c && xe(!0), U(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, E && (s = D);
}
function Fr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function Ir(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, s = e.items, l = Fr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = s.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = s.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (En(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= ie, _ === l) Rr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), zr(e, d, _), zr(e, _, y), Rr(_, y, n), d = _, p = [], m = [], l = Fr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Rr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					zr(e, S.prev, C.next), zr(e, d, S), zr(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), Rr(_, l, n), zr(e, _.prev, _.next), zr(e, _, d === null ? e.effect.first : d.next), zr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = Fr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = Fr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (Mr(e, c(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var ee = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || ee.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && ee.push(l), l = Fr(l.next);
		var te = ee.length;
		if (te > 0) {
			var ne = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < te; v += 1) ee[v].nodes?.a?.measure();
				for (v = 0; v < te; v += 1) ee[v].nodes?.a?.fix();
			}
			jr(e, ee, ne);
		}
	}
	a && Je(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Lr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Lt(n) : /* @__PURE__ */ Rt(n, !1, !1) : null, l = o & 2 ? Lt(i) : null;
	return {
		v: c,
		i: l,
		e: _n(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Rr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Zt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function zr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
function Br(t, n, a = !1, o = !1, s = !1, c = !1) {
	var l = t, u = "";
	if (a) {
		var d = t;
		E && (l = Se(/* @__PURE__ */ Xt(d)));
	}
	B(() => {
		var t = H;
		if (u === (u = n() ?? "")) E && Ce();
		else if (a && !E) t.nodes = null, d.innerHTML = u, u !== "" && gr(/* @__PURE__ */ Xt(d), d.lastChild);
		else if (t.nodes !== null && (Sn(t.nodes.start, t.nodes.end), t.nodes = null), u !== "") {
			if (E) {
				for (var c = D.data, f = Ce(), p = f; f !== null && (f.nodeType !== 8 || f.data !== "");) p = f, f = /* @__PURE__ */ Zt(f);
				if (f === null) throw ye(), e;
				gr(D, p), l = Se(f);
			} else {
				var m = tn(o ? "svg" : s ? "math" : "template", o ? r : s ? i : void 0);
				m.innerHTML = u;
				var h = o || s ? m : m.content;
				if (gr(/* @__PURE__ */ Xt(h), h.lastChild), o || s) for (; /* @__PURE__ */ Xt(h);) l.before(/* @__PURE__ */ Xt(h));
				else l.before(h);
			}
		}
	});
}
// node_modules/svelte/src/internal/shared/attributes.js
var Vr = [..." 	\n\r\f\xA0\v﻿"];
function Hr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Vr.includes(r[o - 1])) && (s === r.length || Vr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Ur(e, t, n, r, i, a) {
	var o = e[pe];
	if (E || o !== n || o === void 0) {
		var s = Hr(n, r, a);
		E && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[pe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Wr = Symbol("is custom element"), Gr = Symbol("is html"), Kr = _e ? "link" : "LINK";
function qr(e) {
	if (E) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					Jr(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					Jr(e, "checked", null), e.checked = r;
				}
			}
		};
		e[ge] = n, Je(n), et();
	}
}
function Jr(e, t, n, r) {
	var i = Yr(e);
	if (!(E && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Kr))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[de] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Zr(e).has(t) ? e[t] = n : (!E || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Yr(e) {
	return e[fe] ??= {
		[Wr]: e.nodeName.includes("-"),
		[Gr]: e.namespaceURI === n
	};
}
var Xr = /* @__PURE__ */ new Map();
function Zr(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Xr.get(t);
	if (n) return n;
	Xr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = d(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = m(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/input.js
function Qr(e, t, n = t) {
	var r = /* @__PURE__ */ new WeakSet();
	nt(e, "input", async (i) => {
		var a = i ? e.defaultValue : e.value;
		if (a = $r(e) ? ei(a) : a, n(a), M !== null && r.add(M), await $n(), a !== (a = t())) {
			var o = e.selectionStart, s = e.selectionEnd, c = e.value.length;
			if (e.value = a ?? "", s !== null) {
				var l = e.value.length;
				o === s && s === c && l > c ? (e.selectionStart = l, e.selectionEnd = l) : (e.selectionStart = o, e.selectionEnd = Math.min(s, l));
			}
		}
	}), (E && e.defaultValue !== e.value || nr(t) == null && e.value) && (n($r(e) ? ei(e.value) : e.value), M !== null && r.add(M)), hn(() => {
		var n = t();
		if (e === document.activeElement) {
			var i = M;
			if (r.has(i)) return;
		}
		$r(e) && n === ei(e.value) || (e.type !== "date" || n || e.value) && n !== e.value && (e.value = n ?? "");
	});
}
function $r(e) {
	var t = e.type;
	return t === "number" || t === "range";
}
function ei(e) {
	return e === "" ? null : +e;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var ti = !1;
function ni(e) {
	var t = ti;
	try {
		return ti = !1, [e(), ti];
	} finally {
		ti = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
function ri(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, l = void 0, d = () => o && i ? (l ??= /* @__PURE__ */ st(r), U(l)) : (c && (c = !1, s = o ? nr(r) : r), s);
	let f;
	if (a) {
		var p = ce in e || ue in e;
		f = u(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = ni(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Fe(t), f(m)));
	var g = i ? () => {
		var n = e[t];
		return n === void 0 ? d() : (c = !0, n);
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
	var v = !1, y = (n & 1 ? st : dt)(() => (v = !1, g()));
	a && U(y);
	var b = H;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? U(y) : i && a ? F(e) : e;
			return P(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return An && v || b.f & 16384 ? y.v : U(y);
	});
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/attention.js
function ii(e, t, n) {
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
async function ai(e, t) {
	if (!await e.confirm(t.ask)) return !1;
	try {
		return await t.run(), e.toast(t.label + ": done"), !0;
	} catch (n) {
		return e.toast(t.label + " failed: " + String(n.message || n)), !1;
	}
}
var oi = Symbol.for("jobshop.desk.attention.pending"), si = "jobshop-attention-pending";
function ci() {
	let e = globalThis;
	return e[oi] ??= /* @__PURE__ */ new Map();
}
function li(e, t) {
	ci().set(e, t), globalThis.dispatchEvent?.(new CustomEvent(si, { detail: {
		key: e,
		label: t
	} }));
}
function ui(e) {
	ci().delete(e);
}
function di(e) {
	let t = (t) => e(t.detail.key, t.detail.label);
	return globalThis.addEventListener?.(si, t), () => globalThis.removeEventListener?.(si, t);
}
// kit/format.js
function fi(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
function pi(e, t) {
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
// node_modules/marked/lib/marked.esm.js
function mi() {
	return {
		async: !1,
		breaks: !1,
		extensions: null,
		gfm: !0,
		hooks: null,
		pedantic: !1,
		renderer: null,
		silent: !1,
		tokenizer: null,
		walkTokens: null
	};
}
var hi = mi();
function gi(e) {
	hi = e;
}
var _i = { exec: () => null };
function vi(e) {
	let t = [];
	return (n) => {
		let r = Math.max(0, Math.min(3, n - 1)), i = t[r];
		return i || (i = e(r), t[r] = i), i;
	};
}
function J(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(Y.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var yi = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), Y = {
	codeRemoveIndent: /^(?: {0,3}\t| {1,4})/gm,
	outputLinkReplace: /\\([\[\]])/g,
	indentCodeCompensation: /^(\s+)(?:```)/,
	beginningSpace: /^\s+/,
	endingHash: /#$/,
	startingSpaceChar: /^ /,
	endingSpaceChar: / $/,
	endingSpaceTabChar: /[ \t]$/,
	nonSpaceChar: /[^ ]/,
	newLineCharGlobal: /\n/g,
	tabCharGlobal: /\t/g,
	leadingSpaceTab: /^[ \t]+/,
	multipleSpaceGlobal: /\s+/g,
	blankLine: /^[ \t]*$/,
	doubleBlankLine: /\n[ \t]*\n[ \t]*$/,
	blockquoteStart: /^ {0,3}>/,
	blockquoteSetextReplace: /\n {0,3}((?:=+|-+)[ \t]*)(?=\n|$)/g,
	blockquoteSetextReplace2: /^ {0,3}>[ \t]?/gm,
	blockquoteMarkerWhitespace: /^( {0,3}>)([ \t]+)/gm,
	listReplaceNesting: /^ {1,4}(?=( {4})*[^ ])/g,
	listIsTask: /^\[[ xX]\][ \t]+\S/,
	listReplaceTask: /^\[[ xX]\][ \t]+/,
	listTaskCheckbox: /\[[ xX]\]/,
	anyLine: /\n.*\n/,
	hrefBrackets: /^<(.*)>$/,
	tableDelimiter: /[:|]/,
	tableAlignChars: /^\||\| *$/g,
	tableRowBlankLine: /\n[ \t]*$/,
	tableAlignRight: /^ *-+: *$/,
	tableAlignCenter: /^ *:-+: *$/,
	tableAlignLeft: /^ *:-+ *$/,
	startATag: /^<a /i,
	endATag: /^<\/a>/i,
	startPreScriptTag: /^<(pre|code|kbd|script)(\s|>)/i,
	endPreScriptTag: /^<\/(pre|code|kbd|script)(\s|>)/i,
	startAngleBracket: /^</,
	endAngleBracket: />$/,
	pedanticHrefTitle: /^([^'"]*[^\s])\s+(['"])(.*)\2/,
	unicodeAlphaNumeric: /[\p{L}\p{N}]/u,
	numericCharacterReference: /&#(?:(\d{1,7})|[Xx]([A-Fa-f0-9]{1,6}));/,
	numericCharacterReferenceOrAmpersand: /&#(?:(\d{1,7})|[Xx]([A-Fa-f0-9]{1,6}));|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
	escapeTest: /[&<>"']/,
	escapeReplace: /[&<>"']/g,
	escapeTestNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/,
	escapeReplaceNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
	caret: /(^|[^\[])\^/g,
	percentDecode: /%25/g,
	findPipe: /\|/g,
	splitPipe: / \|/,
	slashPipe: /\\\|/g,
	carriageReturn: /\r\n|\r/g,
	spaceLine: /^ +$/gm,
	notSpaceStart: /^\S*/,
	endingNewline: /\n$/,
	listItemRegex: (e) => RegExp(`^( {0,3}${e})((?:[	 ][^\\n]*)?(?:\\n|$))`),
	nextBulletRegex: vi((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: vi((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: vi((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: vi((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: vi((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: vi((e) => RegExp(`^ {0,${e}}(?:</?(?:${Pi})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: vi((e) => RegExp(`^ {0,${e}}>`))
}, bi = /^(?:[ \t]*(?:\n|$))+/, xi = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, Si = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, Ci = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, wi = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, Ti = / {0,3}(?:[*+-]|\d{1,9}[.)])/, Ei = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, Di = J(Ei).replace(/bull/g, Ti).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), Oi = J(Ei).replace(/bull/g, Ti).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), ki = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, Ai = /^[^\n]+/, ji = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, Mi = J(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", ji).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), Ni = J(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, Ti).getRegex(), Pi = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Fi = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, Ii = J("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Fi).replace("tag", Pi).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), Li = (e) => J(ki).replace("hr", Ci).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Pi).getRegex(), Ri = Li(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), zi = Li(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), Bi = {
	blockquote: J(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", zi).getRegex(),
	code: xi,
	def: Mi,
	fences: Si,
	heading: wi,
	hr: Ci,
	html: Ii,
	lheading: Di,
	list: Ni,
	newline: bi,
	paragraph: Ri,
	table: _i,
	text: Ai
}, Vi = J("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", Ci).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Pi).getRegex(), Hi = {
	...Bi,
	lheading: Oi,
	table: Vi,
	paragraph: J(ki).replace("hr", Ci).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", Vi).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Pi).getRegex()
}, Ui = {
	...Bi,
	html: J("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", Fi).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: _i,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: J(ki).replace("hr", Ci).replace("heading", " *#{1,6} *[^\n]").replace("lheading", Di).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, Wi = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Gi = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Ki = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, qi = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, Ji = /[\p{P}\p{S}]/u, Yi = /[\s\p{P}\p{S}]/u, Xi = /[^\s\p{P}\p{S}]/u, Zi = J(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, Yi).getRegex(), Qi = /[\p{Pi}\p{Ps}"']/u, $i = /(?!~)[\p{P}\p{S}]/u, ea = /(?!~)[\s\p{P}\p{S}]/u, ta = /(?:[^\s\p{P}\p{S}]|~)/u, na = J(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", yi ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), ra = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, ia = J(ra, "u").replace(/punct/g, Ji).getRegex(), aa = J(ra, "u").replace(/punct/g, $i).getRegex(), oa = J(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, Qi).replace(/punct/g, Ji).getRegex(), sa = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", ca = J(sa, "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), la = J(sa, "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), ua = J("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), da = J("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), fa = J("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), pa = J(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, Ji).getRegex(), ma = J("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), ha = J(/\\(punct)/, "gu").replace(/punct/g, Ji).getRegex(), ga = J(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, Ji).getRegex(), _a = J(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), va = J(Fi).replace("(?:-->|$)", "-->").getRegex(), ya = J("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", va).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), ba = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, xa = J(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", ba).getRegex(), Sa = J(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", xa).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), Ca = J(/^!?\[(label)\]\[(ref)\]/).replace("label", xa).replace("ref", ji).getRegex(), wa = J(/^!?\[(ref)\](?:\[\])?/).replace("ref", ji).getRegex(), Ta = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, Ea = J(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", ba).getRegex(), Da = J("reflink|nolink(?!\\()", "g").replace("reflink", J(/^!?\[(label)\]\[(ref)\]/).replace("label", Ea).replace("ref", Ta).getRegex()).replace("nolink", J(/^!?\[(ref)\](?:\[\])?/).replace("ref", Ta).getRegex()).getRegex(), Oa = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, ka = J(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), Aa = {
	_backpedal: _i,
	anyPunctuation: ha,
	anyPunctuationOutsideCode: ga,
	autolink: _a,
	blockSkip: na,
	br: Ki,
	code: Gi,
	del: _i,
	delLDelim: _i,
	delRDelim: _i,
	emStrongLDelim: ia,
	emStrongRDelimAst: ca,
	emStrongRDelimUnd: da,
	escape: Wi,
	link: Sa,
	nolink: wa,
	punctuation: Zi,
	reflink: Ca,
	reflinkSearch: Da,
	tag: ya,
	text: qi,
	url: _i
}, ja = {
	...Aa,
	emStrongLDelim: oa,
	emStrongRDelimAst: ua,
	emStrongRDelimUnd: fa,
	link: J(/^!?\[(label)\]\((.*?)\)/).replace("label", xa).getRegex(),
	reflink: J(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", xa).getRegex()
}, Ma = {
	...Aa,
	emStrongRDelimAst: la,
	emStrongLDelim: aa,
	delLDelim: pa,
	delRDelim: ma,
	url: J(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", ka).replace("protocol", Oa).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: J(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", Oa).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", yi ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, Na = {
	...Ma,
	br: J(Ki).replace("{2,}", "*").getRegex(),
	text: J(Ma.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, Pa = {
	normal: Bi,
	gfm: Hi,
	pedantic: Ui
}, Fa = {
	normal: Aa,
	gfm: Ma,
	breaks: Na,
	pedantic: ja
}, Ia = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, La = (e) => Ia[e];
function Ra(e, t) {
	if (t) {
		if (Y.escapeTest.test(e)) return e.replace(Y.escapeReplace, La);
	} else if (Y.escapeTestNoEncode.test(e)) return e.replace(Y.escapeReplaceNoEncode, La);
	return e;
}
function za(e) {
	return Y.numericCharacterReference.test(e) ? e.replace(Y.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function Ba(e) {
	try {
		e = encodeURI(e).replace(Y.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function Va(e, t) {
	let n = e.replace(Y.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(Y.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(Y.slashPipe, "|");
	return n;
}
function Ha(e, t, n) {
	let r = e.length;
	if (r === 0) return "";
	let i = 0;
	for (; i < r;) {
		let a = e.charAt(r - i - 1);
		if (a === t && !n) i++;
		else if (a !== t && n) i++;
		else break;
	}
	return e.slice(0, r - i);
}
function Ua(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && Y.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function Wa(e) {
	return e.trim().replace(Y.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function Ga(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Ka = /* @__PURE__ */ new WeakMap();
function qa(e) {
	let t = {
		"*": -1,
		_: -1
	}, n = -1, r = e.length;
	for (let i = 0; i < r; i++) {
		let a = e[i];
		if (a !== "*" && a !== "_" && a !== "~") continue;
		let o = i === 0 ? "\n" : e[i - 1], s = i + 1;
		for (; s < r && e[s] === a;) s++;
		o !== " " && o !== "	" && o !== "\n" && o !== "\r" && o !== "\f" && (a === "~" ? n = i : t[a] = i), i = s - 1;
	}
	return {
		maskedSrc: e,
		lastPossibleEmStrongCloserIndex: t,
		lastPossibleDelCloserIndex: n
	};
}
function Ja(e, t) {
	let n = qa(t), r = Ka.get(e);
	r ? r.push(n) : Ka.set(e, [n]);
}
function Ya(e) {
	let t = Ka.get(e);
	t && (t.pop(), t.length === 0 && Ka.delete(e));
}
function Xa(e, t, n) {
	let r = Ka.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : qa(t).lastPossibleEmStrongCloserIndex[n];
}
function Za(e, t) {
	let n = Ka.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : qa(t).lastPossibleDelCloserIndex;
}
function Qa(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function $a(e, t, n, r, i) {
	let a = t.href, o = t.title || null, s = e[1].replace(i.other.outputLinkReplace, "$1"), c = e[0].charAt(0) === "!", l = r.state.inLink, u = r.state.linkEmitted, d = r.state.inRawBlock;
	r.state.inLink = !0, r.state.linkEmitted = !1;
	let f = r.inlineTokens(s), p = r.state.linkEmitted;
	if (r.state.linkEmitted = u, r.state.inLink = l, !c) {
		if (p) {
			r.state.inRawBlock = d;
			return;
		}
		r.state.linkEmitted = !0;
	}
	return {
		type: c ? "image" : "link",
		raw: n,
		href: a,
		title: o,
		text: s,
		tokens: f
	};
}
function eo(e, t, n) {
	let r = e.match(n.other.indentCodeCompensation);
	if (r === null) return t;
	let i = r[1];
	return t.split("\n").map((e) => {
		let t = e.match(n.other.beginningSpace);
		if (t === null) return e;
		let [r] = t;
		return e.slice(Math.min(r.length, i.length));
	}).join("\n");
}
function to(e, t, n, r) {
	if (!t.includes("<")) return !1;
	for (let i = 0; i < t.length; i++) {
		if (t[i] === "\\") {
			i++;
			continue;
		}
		if (t[i] === "`") {
			let e = r.inline.code.exec(t.slice(i));
			if (e) {
				i += e[0].length - 1;
				continue;
			}
		}
		if (t[i] !== "<") continue;
		let a = e.slice(n + i), o = r.inline.tag.exec(a) || r.inline.autolink.exec(a);
		if (o) {
			if (o[0].length > t.length - i) return !0;
			i += o[0].length - 1;
		}
	}
	return !1;
}
var no = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || hi;
	}
	space(e) {
		let t = this.rules.block.newline.exec(e);
		if (t && t[0].length > 0) return {
			type: "space",
			raw: t[0]
		};
	}
	code(e) {
		let t = this.rules.block.code.exec(e);
		if (t) {
			let e = this.options.pedantic ? t[0] : Ua(t[0]);
			return {
				type: "code",
				raw: e,
				codeBlockStyle: "indented",
				text: e.replace(this.rules.other.codeRemoveIndent, "")
			};
		}
	}
	fences(e) {
		let t = this.rules.block.fences.exec(e);
		if (t) {
			let e = t[0], n = eo(e, t[3] || "", this.rules);
			return {
				type: "code",
				raw: e,
				lang: t[2] ? t[2].trim().replace(this.rules.inline.anyPunctuation, "$1") : t[2],
				text: n
			};
		}
	}
	heading(e) {
		let t = this.rules.block.heading.exec(e);
		if (t) {
			let e = t[2].trim();
			if (this.rules.other.endingHash.test(e)) {
				let t = Ha(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: Ha(t[0], "\n"),
				depth: t[1].length,
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	hr(e) {
		let t = this.rules.block.hr.exec(e);
		if (t) return {
			type: "hr",
			raw: Ha(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = Ha(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + Qa(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
				n = n ? `${n}
${s}` : s, r = r ? `${r}
${c}` : c;
				let l = this.lexer.state.top;
				if (this.lexer.state.top = !0, this.lexer.blockTokens(c, i, !0), this.lexer.state.top = l, e.length === 0) break;
				let u = i.at(-1);
				if (u?.type === "code" || u?.type === "heading" || u?.type === "hr" || u?.type === "html" || u?.type === "table") break;
				if (u?.type === "blockquote") {
					let t = u, a = e.join("\n"), o = t.raw + "\n" + a.replace(this.rules.other.blockquoteSetextReplace2, ""), s = this.blockquote(o);
					i[i.length - 1] = s;
					let c = o.substring(s.raw.length).replace(/^\n/, ""), l = c ? c.split("\n").length : 0, d = l ? e.slice(0, -l) : e;
					if (d.length > 0 && (n = `${n}
${d.join("\n")}`), r = r.substring(0, r.length - t.text.length) + s.text, e = e.slice(d.length), e.length > 0 && this.rules.other.blockquoteStart.test(e[0])) continue;
					break;
				}
				if (u?.type === "list") {
					let t = u, a = t.raw + "\n" + e.join("\n"), o = this.list(a);
					i[i.length - 1] = o;
					let s = o.raw.length > t.raw.length ? a.slice(t.raw.length + 1, o.raw.length).split("\n").length : 0;
					if (s > 0) {
						let t = e.slice(0, s).join("\n");
						n = `${n}
${t}`, r = `${r}
${t}`;
					}
					e = e.slice(s);
				}
			}
			return {
				type: "blockquote",
				raw: n,
				tokens: i,
				text: r
			};
		}
	}
	list(e) {
		let t = this.rules.block.list.exec(e);
		if (t) {
			let n = e, r = () => n.length - e.length, i = t[1].trim(), a = i.length > 1, o = {
				type: "list",
				raw: "",
				ordered: a,
				start: a ? +i.slice(0, -1) : "",
				loose: !1,
				items: []
			};
			i = a ? `\\d{1,9}\\${i.slice(-1)}` : `\\${i}`, this.options.pedantic && (i = a ? i : "[*+-]");
			let s = this.rules.other.listItemRegex(i), c = !1;
			for (; e;) {
				let i = !1, a = "";
				if (!(t = s.exec(e)) || this.rules.block.hr.test(e)) break;
				let l = r();
				e = e.substring(t[0].length);
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? Qa(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => Qa(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => Qa(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
						if (l.search(this.rules.other.nonSpaceChar) >= h || !p.trim()) a += "\n" + l.slice(h);
						else {
							if (m || f.replace(this.rules.other.tabCharGlobal, "    ").search(this.rules.other.nonSpaceChar) >= 4 || r.test(f) || i.test(f) || o.test(f) || n.test(f)) break;
							let e = !this.options.pedantic && /^ {0,3}(?:=+|-+) *$/.test(p) && this.rules.block.lheading.test(`${f}
${p}`) ? `    ${l.replace(/^ {0,3}/, "")}` : p;
							a += "\n" + e;
						}
						m = !p.trim(), e = e.substring(c.length + 1), f = l.slice(h);
					}
				}
				let g = n.slice(l, r());
				o.loose || (c ? o.loose = !0 : this.rules.other.doubleBlankLine.test(g) && (c = !0)), o.items.push({
					type: "list_item",
					raw: g,
					task: !!this.options.gfm && this.rules.other.listIsTask.test(a),
					loose: !1,
					text: a,
					tokens: []
				});
			}
			let l = o.items.at(-1);
			if (l) l.raw = l.raw.trimEnd(), l.text = l.text.trimEnd();
			else return;
			o.raw = n.slice(0, r()).trimEnd();
			for (let e of o.items) if (this.lexer.state.top = !1, e.tokens = this.lexer.blockTokens(e.text, []), !o.loose) {
				let t = e.tokens.filter((e) => e.type === "space");
				o.loose = t.length > 0 && t.some((e) => this.rules.other.anyLine.test(e.raw));
			}
			for (let e of o.items) {
				let t = e.tokens[0];
				if (e.task && (t?.type === "text" || t?.type === "paragraph")) {
					e.text = e.text.replace(this.rules.other.listReplaceTask, ""), t.raw = t.raw.replace(this.rules.other.listReplaceTask, ""), t.text = t.text.replace(this.rules.other.listReplaceTask, "");
					for (let e = this.lexer.inlineQueue.length - 1; e >= 0; e--) if (this.rules.other.listIsTask.test(this.lexer.inlineQueue[e].src)) {
						this.lexer.inlineQueue[e].src = this.lexer.inlineQueue[e].src.replace(this.rules.other.listReplaceTask, "");
						break;
					}
					let n = this.rules.other.listTaskCheckbox.exec(e.raw);
					if (n) {
						let t = {
							type: "checkbox",
							raw: n[0] + " ",
							checked: n[0] !== "[ ]"
						};
						e.checked = t.checked, o.loose ? e.tokens[0] && ["paragraph", "text"].includes(e.tokens[0].type) && "tokens" in e.tokens[0] && e.tokens[0].tokens ? (e.tokens[0].raw = t.raw + e.tokens[0].raw, e.tokens[0].text = t.raw + e.tokens[0].text, e.tokens[0].tokens.unshift(t)) : e.tokens.unshift({
							type: "paragraph",
							raw: t.raw,
							text: t.raw,
							tokens: [t]
						}) : e.tokens.unshift(t);
					}
				} else e.task &&= !1;
			}
			if (o.loose) for (let e of o.items) {
				e.loose = !0;
				for (let t of e.tokens) t.type === "text" && (t.type = "paragraph");
			}
			return o;
		}
	}
	html(e) {
		let t = this.rules.block.html.exec(e);
		if (t) {
			let e = Ua(t[0]);
			return {
				type: "html",
				block: !0,
				raw: e,
				pre: t[1] === "pre" || t[1] === "script" || t[1] === "style",
				text: e
			};
		}
	}
	def(e) {
		let t = this.rules.block.def.exec(e);
		if (t) {
			if (!this.rules.other.startAngleBracket.test(t[2]) && Ga(t[2], "()") !== -1) return;
			let e = Wa(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: Ha(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = Va(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: Ha(t[0], "\n"),
			header: [],
			align: [],
			rows: []
		};
		if (n.length === r.length) {
			for (let e of r) this.rules.other.tableAlignRight.test(e) ? a.align.push("right") : this.rules.other.tableAlignCenter.test(e) ? a.align.push("center") : this.rules.other.tableAlignLeft.test(e) ? a.align.push("left") : a.align.push(null);
			for (let e = 0; e < n.length; e++) a.header.push({
				text: n[e],
				tokens: this.lexer.inline(n[e]),
				header: !0,
				align: a.align[e]
			});
			for (let e of i) a.rows.push(Va(e, a.header.length).map((e, t) => ({
				text: e,
				tokens: this.lexer.inline(e),
				header: !1,
				align: a.align[t]
			})));
			return a;
		}
	}
	lheading(e) {
		let t = this.rules.block.lheading.exec(e);
		if (t) {
			let e = t[1].trim();
			return {
				type: "heading",
				raw: Ha(t[0], "\n"),
				depth: t[2].charAt(0) === "=" ? 1 : 2,
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	paragraph(e) {
		let t = this.rules.block.paragraph.exec(e);
		if (t) {
			let e = t[1].charAt(t[1].length - 1) === "\n" ? t[1].slice(0, -1) : t[1];
			return {
				type: "paragraph",
				raw: t[0],
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	text(e) {
		let t = this.rules.block.text.exec(e);
		if (t) return {
			type: "text",
			raw: t[0],
			text: t[0],
			tokens: this.lexer.inline(t[0])
		};
	}
	escape(e) {
		let t = this.rules.inline.escape.exec(e);
		if (t) return {
			type: "escape",
			raw: t[0],
			text: t[1]
		};
	}
	tag(e) {
		let t = this.rules.inline.tag.exec(e);
		if (t) return !this.lexer.state.inLink && this.rules.other.startATag.test(t[0]) ? this.lexer.state.inLink = !0 : this.lexer.state.inLink && this.rules.other.endATag.test(t[0]) && (this.lexer.state.inLink = !1), !this.lexer.state.inRawBlock && this.rules.other.startPreScriptTag.test(t[0]) ? this.lexer.state.inRawBlock = !0 : this.lexer.state.inRawBlock && this.rules.other.endPreScriptTag.test(t[0]) && (this.lexer.state.inRawBlock = !1), {
			type: "html",
			raw: t[0],
			inLink: this.lexer.state.inLink,
			inRawBlock: this.lexer.state.inRawBlock,
			block: !1,
			text: t[0]
		};
	}
	link(e) {
		if (this.lexer.state.linkParenPossible === !1) return;
		let t = this.rules.inline.link.exec(e);
		if (t) {
			let n = t[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && to(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = Ha(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = Ga(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = Ha(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), $a(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && to(e, n[1], r, this.rules)) return;
			let i = t[Wa(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return $a(n, i, n[0], this.lexer, this.rules);
		}
	}
	emStrong(e, t, n = "") {
		let r = this.rules.inline.emStrongLDelim.exec(e);
		if (!r || !r[1] && !r[2] && !r[3] && !r[4]) return;
		let i = r[0][0], a = n === i, o = t.length - e.length, s = o;
		for (; !this.options.pedantic && s > 0 && t[s - 1] === i;) s--;
		let c = [...t.slice(Math.max(0, s - 2), s)].at(-1) ?? "", l = c === "*" || c === "_" || !!this.rules.inline.punctuation.exec(c);
		s < o && (n = c);
		let u = s < o ? l : !!this.rules.inline.punctuation.exec(n);
		if (r[4] && n.match(this.rules.other.unicodeAlphaNumeric)) return;
		let d = r[1] || r[3] || "";
		if (!d || !n || u) {
			let n = [...r[0]].length - 1, u, f, p = n, m = 0, h = o - s + n, g = !this.options.pedantic && !!c.trim() && (!!d || i === "*" && !l);
			if (i !== "*" && i !== "_") return;
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = Xa(this, t, i), y = t.length - e.length + n;
			if (v < y) return;
			t = t.slice(-1 * e.length + n);
			let b = v - y;
			for (_.lastIndex = 0; (r = _.exec(t)) !== null && !(r.index > b);) {
				if (u = r[1] || r[2] || r[3] || r[4] || r[5] || r[6], !u) continue;
				if (f = [...u].length, r[3] || r[4]) {
					p += f;
					continue;
				}
				if (r[5] || r[6] || g) {
					if (a && (r[5] || r[6])) break;
					if (h % 3 && !((h + f) % 3)) {
						m += f;
						continue;
					}
				}
				if (p -= f, p > 0) continue;
				f = Math.min(f, f + p + m);
				let t = [...r[0]][0].length, i = e.slice(0, n + r.index + t + f);
				if (Math.min(n, f) % 2) {
					let e = i.slice(1, -1);
					return {
						type: "em",
						raw: i,
						text: e,
						tokens: this.lexer.inlineTokens(e)
					};
				}
				let o = i.slice(2, -2);
				return {
					type: "strong",
					raw: i,
					text: o,
					tokens: this.lexer.inlineTokens(o)
				};
			}
		}
	}
	codespan(e) {
		let t = this.rules.inline.code.exec(e);
		if (t) {
			let e = t[2].replace(this.rules.other.newLineCharGlobal, " "), n = this.rules.other.nonSpaceChar.test(e), r = this.rules.other.startingSpaceChar.test(e) && this.rules.other.endingSpaceChar.test(e);
			return n && r && (e = e.substring(1, e.length - 1)), {
				type: "codespan",
				raw: t[0],
				text: e
			};
		}
	}
	br(e) {
		let t = this.rules.inline.br.exec(e);
		if (t) return {
			type: "br",
			raw: t[0]
		};
	}
	del(e, t, n = "") {
		let r = this.rules.inline.delLDelim.exec(e);
		if (r && (!r[1] || !n || this.rules.inline.punctuation.exec(n))) {
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = Za(this, t), l = t.length - e.length + n;
			if (c < l) return;
			t = t.slice(-1 * e.length + n);
			let u = c - l;
			for (s.lastIndex = 0; (r = s.exec(t)) !== null && !(r.index > u);) {
				if (i = r[1] || r[2] || r[3] || r[4] || r[5] || r[6], !i || (a = [...i].length, a !== n)) continue;
				if (r[3] || r[4]) {
					o += a;
					continue;
				}
				if (o -= a, o > 0) continue;
				a = Math.min(a, a + o);
				let t = [...r[0]][0].length, s = e.slice(0, n + r.index + t + a), c = s.slice(n, -n);
				return {
					type: "del",
					raw: s,
					text: c,
					tokens: this.lexer.inlineTokens(c)
				};
			}
		}
	}
	autolink(e) {
		let t = this.rules.inline.autolink.exec(e);
		if (t) {
			let e, n;
			return t[2] === "@" ? (e = t[1], n = "mailto:" + e) : (e = t[1], n = e), {
				type: "link",
				raw: t[0],
				text: e,
				href: n,
				autolink: !0,
				tokens: [{
					type: "text",
					raw: e,
					text: e
				}]
			};
		}
	}
	url(e) {
		let t;
		if (t = this.rules.inline.url.exec(e)) {
			let e, n;
			if (t[2] === "@") e = t[0], n = "mailto:" + e;
			else {
				let r;
				do
					r = t[0], t[0] = this.rules.inline._backpedal.exec(t[0])?.[0] ?? "";
				while (r !== t[0]);
				e = t[0], n = t[1] === "www." ? "http://" + t[0] : t[0];
			}
			return {
				type: "link",
				raw: t[0],
				text: e,
				href: n,
				autolink: !0,
				tokens: [{
					type: "text",
					raw: e,
					text: e
				}]
			};
		}
	}
	inlineText(e) {
		let t = this.rules.inline.text.exec(e);
		if (t) {
			let e = this.lexer.state.inRawBlock;
			return {
				type: "text",
				raw: t[0],
				text: e ? t[0] : za(t[0]),
				escaped: e
			};
		}
	}
}, ro = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || hi, this.options.tokenizer = this.options.tokenizer || new no(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: Y,
			block: Pa.normal,
			inline: Fa.normal
		};
		this.options.pedantic ? (t.block = Pa.pedantic, t.inline = Fa.pedantic) : this.options.gfm && (t.block = Pa.gfm, t.inline = this.options.breaks ? Fa.breaks : Fa.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: Pa,
			inline: Fa
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(Y.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(Y.tabCharGlobal, "    ").replace(Y.spaceLine, ""));
		let r = 1 / 0;
		for (; e;) {
			if (e.length < r) r = e.length;
			else {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
			let i;
			if (this.options.extensions?.block?.some((n) => (i = n.call({ lexer: this }, e, t)) ? (e = e.substring(i.raw.length), t.push(i), !0) : !1)) continue;
			if (i = this.tokenizer.space(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				i.raw.length === 1 && n !== void 0 ? n.raw += "\n" : t.push(i);
				continue;
			}
			if (i = this.tokenizer.code(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.text, this.inlineQueue.at(-1).src = n.text) : t.push(i);
				continue;
			}
			if ((i = this.tokenizer.fences(e)) || (i = this.tokenizer.heading(e)) || (i = this.tokenizer.hr(e)) || (i = this.tokenizer.blockquote(e)) || (i = this.tokenizer.list(e)) || (i = this.tokenizer.html(e))) {
				e = e.substring(i.raw.length), t.push(i);
				continue;
			}
			if (i = this.tokenizer.def(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.raw, this.inlineQueue.at(-1).src = n.text) : (this.tokens.links[i.tag] || (this.tokens.links[i.tag] = {
					href: i.href,
					title: i.title
				}), t.push(i));
				continue;
			}
			if ((i = this.tokenizer.table(e)) || (i = this.tokenizer.lheading(e))) {
				e = e.substring(i.raw.length), t.push(i);
				continue;
			}
			let a = e;
			if (this.options.extensions?.startBlock) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startBlock.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (a = e.substring(0, t + 1));
			}
			if (this.state.top && (i = this.tokenizer.paragraph(a))) {
				let r = t.at(-1);
				n && r?.type === "paragraph" ? (r.raw += (r.raw.endsWith("\n") ? "" : "\n") + i.raw, r.text += "\n" + i.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = r.text) : t.push(i), n = a.length !== e.length, e = e.substring(i.raw.length);
			} else if (i = this.tokenizer.text(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = n.text) : t.push(i);
			} else if (e) {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
		}
		return this.state.top = !0, t;
	}
	inline(e, t = []) {
		return this.inlineQueue.push({
			src: e,
			tokens: t
		}), t;
	}
	linkInText(e) {
		if (!e.includes("[")) return !1;
		let t = this.tokenizer.rules.inline.link;
		for (let n of e.matchAll(this.tokenizer.rules.inline.blockSkip)) if (t.test(n[0]) && e.charAt(n.index - 1) !== "!") return !0;
		for (let t of e.matchAll(this.tokenizer.rules.inline.reflinkSearch)) {
			let e = t[0], n = e.lastIndexOf("[");
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, Wa(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
		}
		return !1;
	}
	inlineTokens(e, t = []) {
		this.tokenizer.lexer = this;
		let n = this.state.linkParenPossible;
		this.state.linkParenPossible = n && e.includes(")");
		try {
			return this.#e(e, t);
		} finally {
			this.state.linkParenPossible = n;
		}
	}
	#e(e, t) {
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== no.prototype.emStrong || this.tokenizer.del !== no.prototype.del ? this.#n(e) : e;
		Ja(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			Ya(this.tokenizer);
		}
	}
	#t(e, t, n) {
		let r = !1, i = "", a = 1 / 0;
		for (; e;) {
			if (e.length < a) a = e.length;
			else {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
			r || (i = ""), r = !1;
			let o;
			if (this.options.extensions?.inline?.some((t) => (o = t.call({ lexer: this }, e, n)) ? (e = e.substring(o.raw.length), n.push(o), !0) : !1)) continue;
			if ((o = this.tokenizer.escape(e)) || (o = this.tokenizer.tag(e)) || (o = this.tokenizer.link(e))) {
				e = e.substring(o.raw.length), n.push(o);
				continue;
			}
			if (o = this.tokenizer.reflink(e, this.tokens.links)) {
				e = e.substring(o.raw.length);
				let t = n.at(-1);
				o.type === "text" && t?.type === "text" ? (t.raw += o.raw, t.text += o.text) : n.push(o);
				continue;
			}
			if ((o = this.tokenizer.emStrong(e, t, i)) || (o = this.tokenizer.codespan(e)) || (o = this.tokenizer.br(e)) || (o = this.tokenizer.del(e, t, i)) || (o = this.tokenizer.autolink(e)) || !this.state.inLink && (o = this.tokenizer.url(e))) {
				e = e.substring(o.raw.length), n.push(o);
				continue;
			}
			let s = e;
			if (this.options.extensions?.startInline) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startInline.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (s = e.substring(0, t + 1));
			}
			if (o = this.tokenizer.inlineText(s)) {
				e = e.substring(o.raw.length), o.raw.slice(-1) !== "_" && (i = o.raw.slice(-1)), r = !0;
				let t = n.at(-1);
				t?.type === "text" ? (t.raw += o.raw, t.text += o.text) : n.push(o);
			} else if (e) {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
		}
		return n;
	}
	#n(e) {
		let t = e;
		if (this.tokens.links && e.includes("[")) {
			let e = this.tokenizer.rules.inline.reflinkSearch, n = (t) => {
				let r = t.lastIndexOf("[");
				if (!Object.hasOwn(this.tokens.links, Wa(t.slice(r + 1, -1)))) return t;
				if (r > 1 && t.charAt(0) !== "!") {
					let i = t.slice(1, r - 1);
					if (this.linkInText(i)) return "[" + i.replace(e, n) + "][" + "a".repeat(t.length - r - 2) + "]";
				}
				return "[" + "a".repeat(t.length - 2) + "]";
			};
			t = t.replace(e, n);
		}
		return t = t.replace(this.tokenizer.rules.inline.anyPunctuationOutsideCode, (e, t, n) => n ? "+".repeat(e.length) : t ? t + e.slice(t.length, -t.length).replace(/`/g, "+") + t : e), t = t.replace(this.tokenizer.rules.inline.blockSkip, (e, t, n) => {
			let r = n ? n.length : 0;
			return e.slice(0, r) + "[" + "a".repeat(e.length - r - 2) + "]";
		}), t = this.options.hooks?.emStrongMask?.call({ lexer: this }, t) ?? t, t;
	}
	infiniteLoopError(e) {
		let t = "Infinite loop on byte: " + e;
		if (this.options.silent) console.error(t);
		else throw Error(t);
	}
}, io = class {
	options;
	parser;
	constructor(e) {
		this.options = e || hi;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(Y.notSpaceStart)?.[0], i = e ? e.replace(Y.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + Ra(r) + "\">" + (n ? i : Ra(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : Ra(i, !0)) + "</code></pre>\n";
	}
	blockquote({ tokens: e }) {
		return `<blockquote>
${this.parser.parse(e)}</blockquote>
`;
	}
	html({ text: e }) {
		return e;
	}
	def(e) {
		return "";
	}
	heading({ tokens: e, depth: t }) {
		return `<h${t}>${this.parser.parseInline(e)}</h${t}>
`;
	}
	hr(e) {
		return "<hr>\n";
	}
	list(e) {
		let t = e.ordered, n = e.start, r = "";
		for (let t = 0; t < e.items.length; t++) {
			let n = e.items[t];
			r += this.listitem(n);
		}
		let i = t ? "ol" : "ul", a = t && n !== 1 ? " start=\"" + n + "\"" : "";
		return "<" + i + a + ">\n" + r + "</" + i + ">\n";
	}
	listitem(e) {
		return `<li>${this.parser.parse(e.tokens)}</li>
`;
	}
	checkbox({ checked: e }) {
		return "<input " + (e ? "checked=\"\" " : "") + "disabled=\"\" type=\"checkbox\"> ";
	}
	paragraph({ tokens: e }) {
		return `<p>${this.parser.parseInline(e)}</p>
`;
	}
	table(e) {
		let t = "", n = "";
		for (let t = 0; t < e.header.length; t++) n += this.tablecell(e.header[t]);
		t += this.tablerow({ text: n });
		let r = "";
		for (let t = 0; t < e.rows.length; t++) {
			let i = e.rows[t];
			n = "";
			for (let e = 0; e < i.length; e++) n += this.tablecell(i[e]);
			r += this.tablerow({ text: n });
		}
		return r &&= `<tbody>${r}</tbody>`, "<table>\n<thead>\n" + t + "</thead>\n" + r + "</table>\n";
	}
	tablerow({ text: e }) {
		return `<tr>
${e}</tr>
`;
	}
	tablecell(e) {
		let t = this.parser.parseInline(e.tokens), n = e.header ? "th" : "td";
		return (e.align ? `<${n} align="${e.align}">` : `<${n}>`) + t + `</${n}>
`;
	}
	strong({ tokens: e }) {
		return `<strong>${this.parser.parseInline(e)}</strong>`;
	}
	em({ tokens: e }) {
		return `<em>${this.parser.parseInline(e)}</em>`;
	}
	codespan({ text: e }) {
		return `<code>${Ra(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? Ra(n, !0) : this.parser.parseInline(r), o = Ba(e);
		if (o === null) return a;
		e = Ra(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + Ra(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = Ba(e);
		if (i === null) return Ra(n);
		e = i;
		let a = `<img src="${Ra(e)}" alt="${Ra(n)}"`;
		return t && (a += ` title="${Ra(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : Ra(e.text);
	}
}, ao = class {
	parser;
	strong({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	em({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	codespan({ text: e }) {
		return e;
	}
	del({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	html({ text: e }) {
		return e;
	}
	text({ text: e }) {
		return e;
	}
	link({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	image({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	br() {
		return "";
	}
	checkbox({ raw: e }) {
		return e;
	}
}, oo = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || hi, this.options.renderer = this.options.renderer || new io(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new ao(), this.textRenderer.parser = this;
	}
	static parse(t, n) {
		return new e(n).parse(t);
	}
	static parseInline(t, n) {
		return new e(n).parseInline(t);
	}
	parse(e) {
		this.renderer.parser = this;
		let t = "";
		for (let n = 0; n < e.length; n++) {
			let r = e[n];
			if (this.options.extensions?.renderers?.[r.type]) {
				let e = r, n = this.options.extensions.renderers[e.type].call({ parser: this }, e);
				if (n !== !1 || ![
					"space",
					"hr",
					"heading",
					"code",
					"table",
					"blockquote",
					"list",
					"checkbox",
					"html",
					"def",
					"paragraph",
					"text"
				].includes(e.type)) {
					t += n || "";
					continue;
				}
			}
			let i = r;
			switch (i.type) {
				case "space":
					t += this.renderer.space(i);
					break;
				case "hr":
					t += this.renderer.hr(i);
					break;
				case "heading":
					t += this.renderer.heading(i);
					break;
				case "code":
					t += this.renderer.code(i);
					break;
				case "table":
					t += this.renderer.table(i);
					break;
				case "blockquote":
					t += this.renderer.blockquote(i);
					break;
				case "list":
					t += this.renderer.list(i);
					break;
				case "checkbox":
					t += this.renderer.checkbox(i);
					break;
				case "html":
					t += this.renderer.html(i);
					break;
				case "def":
					t += this.renderer.def(i);
					break;
				case "paragraph":
					t += this.renderer.paragraph(i);
					break;
				case "text":
					t += this.renderer.text(i);
					break;
				default: {
					let e = "Token with \"" + i.type + "\" type was not found.";
					if (this.options.silent) return console.error(e), "";
					throw Error(e);
				}
			}
		}
		return t;
	}
	parseInline(e, t = this.renderer) {
		this.renderer.parser = this;
		let n = "";
		for (let r = 0; r < e.length; r++) {
			let i = e[r];
			if (this.options.extensions?.renderers?.[i.type]) {
				let e = this.options.extensions.renderers[i.type].call({ parser: this }, i);
				if (e !== !1 || ![
					"escape",
					"html",
					"link",
					"image",
					"checkbox",
					"strong",
					"em",
					"codespan",
					"br",
					"del",
					"text"
				].includes(i.type)) {
					n += e || "";
					continue;
				}
			}
			let a = i;
			switch (a.type) {
				case "escape":
					n += t.text(a);
					break;
				case "html":
					n += t.html(a);
					break;
				case "link":
					n += t.link(a);
					break;
				case "image":
					n += t.image(a);
					break;
				case "checkbox":
					n += t.checkbox(a);
					break;
				case "strong":
					n += t.strong(a);
					break;
				case "em":
					n += t.em(a);
					break;
				case "codespan":
					n += t.codespan(a);
					break;
				case "br":
					n += t.br(a);
					break;
				case "del":
					n += t.del(a);
					break;
				case "text":
					n += t.text(a);
					break;
				default: {
					let e = "Token with \"" + a.type + "\" type was not found.";
					if (this.options.silent) return console.error(e), "";
					throw Error(e);
				}
			}
		}
		return n;
	}
}, so = class {
	options;
	block;
	constructor(e) {
		this.options = e || hi;
	}
	static passThroughHooks = /* @__PURE__ */ new Set([
		"preprocess",
		"postprocess",
		"processAllTokens",
		"emStrongMask"
	]);
	static passThroughHooksRespectAsync = /* @__PURE__ */ new Set([
		"preprocess",
		"postprocess",
		"processAllTokens"
	]);
	preprocess(e) {
		return e;
	}
	postprocess(e) {
		return e;
	}
	processAllTokens(e) {
		return e;
	}
	emStrongMask(e) {
		return e;
	}
	provideLexer(e = this.block) {
		return e ? ro.lex : ro.lexInline;
	}
	provideParser(e = this.block) {
		return e ? oo.parse : oo.parseInline;
	}
};
function co(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var lo = new class {
	defaults = mi();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = oo;
	Renderer = io;
	TextRenderer = ao;
	Lexer = ro;
	Tokenizer = no;
	Hooks = so;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (co(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) co(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) co(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				co(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					co(n, this.walkTokens(i, t));
				}) : e.tokens && co(n, this.walkTokens(e.tokens, t));
			}
		}
		return n;
	}
	use(...e) {
		let t = this.defaults.extensions || {
			renderers: Object.create(null),
			childTokens: Object.create(null)
		};
		return e.forEach((e) => {
			let n = { ...e };
			if (n.async = this.defaults.async || n.async || !1, e.extensions && (e.extensions.forEach((e) => {
				if (!e.name) throw Error("extension name required");
				if ("renderer" in e) {
					let n = t.renderers[e.name];
					n ? t.renderers[e.name] = function(...t) {
						let r = e.renderer.apply(this, t);
						return r === !1 && (r = n.apply(this, t)), r;
					} : t.renderers[e.name] = e.renderer;
				}
				if ("tokenizer" in e) {
					if (!e.level || e.level !== "block" && e.level !== "inline") throw Error("extension level must be 'block' or 'inline'");
					let n = t[e.level];
					n ? n.unshift(e.tokenizer) : t[e.level] = [e.tokenizer], e.start && (e.level === "block" ? t.startBlock ? t.startBlock.push(e.start) : t.startBlock = [e.start] : e.level === "inline" && (t.startInline ? t.startInline.push(e.start) : t.startInline = [e.start]));
				}
				"childTokens" in e && e.childTokens && (t.childTokens[e.name] = e.childTokens);
			}), n.extensions = t), e.renderer) {
				let t = this.defaults.renderer || new io(this.defaults);
				for (let n in e.renderer) {
					if (!(n in t)) throw Error(`renderer '${n}' does not exist`);
					if (["options", "parser"].includes(n)) continue;
					let r = n, i = e.renderer[r], a = t[r];
					t[r] = (...e) => {
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n === !1 ? "" : n ?? "";
					};
				}
				n.renderer = t;
			}
			if (e.tokenizer) {
				let t = this.defaults.tokenizer || new no(this.defaults);
				for (let n in e.tokenizer) {
					if (!(n in t)) throw Error(`tokenizer '${n}' does not exist`);
					if ([
						"options",
						"rules",
						"lexer"
					].includes(n)) continue;
					let r = n, i = e.tokenizer[r], a = t[r];
					t[r] = (...e) => {
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n;
					};
				}
				n.tokenizer = t;
			}
			if (e.hooks) {
				let t = this.defaults.hooks || new so();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = so.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && so.passThroughHooksRespectAsync.has(n)) return (async () => {
							let n = await i.call(t, e);
							return a.call(t, n);
						})();
						let r = i.call(t, e);
						return a.call(t, r);
					} : (...e) => {
						if (t.options.async) return (async () => {
							let n = await i.apply(t, e);
							return n === !1 && (n = await a.apply(t, e)), n;
						})();
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n;
					};
				}
				n.hooks = t;
			}
			if (e.walkTokens) {
				let t = this.defaults.walkTokens, r = e.walkTokens;
				n.walkTokens = function(e) {
					let n = [];
					return co(n, r.call(this, e)), t && co(n, t.call(this, e)), n;
				};
			}
			this.defaults = {
				...this.defaults,
				...n
			};
		}), this;
	}
	setOptions(e) {
		return this.defaults = {
			...this.defaults,
			...e
		}, this;
	}
	lexer(e, t) {
		return ro.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return oo.parse(e, t ?? this.defaults);
	}
	parseMarkdown(e) {
		return (t, n) => {
			let r = { ...n }, i = {
				...this.defaults,
				...r
			}, a = this.onError(!!i.silent, !!i.async);
			if (this.defaults.async === !0 && r.async === !1) return a(/* @__PURE__ */ Error("marked(): The async option was set to true by an extension. Remove async: false from the parse options object to return a Promise."));
			if (typeof t > "u" || t === null) return a(/* @__PURE__ */ Error("marked(): input parameter is undefined or null"));
			if (typeof t != "string") return a(/* @__PURE__ */ Error("marked(): input parameter is of type " + Object.prototype.toString.call(t) + ", string expected"));
			if (i.hooks && (i.hooks.options = i, i.hooks.block = e), i.async) return (async () => {
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? ro.lex : ro.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? oo.parse : oo.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? ro.lex : ro.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? oo.parse : oo.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + Ra(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function X(e, t) {
	return lo.parse(e, t);
}
X.options = X.setOptions = function(e) {
	return lo.setOptions(e), X.defaults = lo.defaults, gi(X.defaults), X;
}, X.getDefaults = mi, X.defaults = hi;
function uo(...e) {
	return lo.use(...e), X.defaults = lo.defaults, gi(X.defaults), X;
}
X.use = uo, X.walkTokens = function(e, t) {
	return lo.walkTokens(e, t);
}, X.parseInline = lo.parseInline, X.Parser = oo, X.parser = oo.parse, X.Renderer = io, X.TextRenderer = ao, X.Lexer = ro, X.lexer = ro.lex, X.Tokenizer = no, X.Hooks = so, X.parse = X, X.options, X.setOptions, X.walkTokens, X.parseInline, oo.parse, ro.lex;
// node_modules/dompurify/dist/purify.es.mjs
function fo(e, t) {
	this.v = e, this.k = t;
}
function po(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function mo(e) {
	if (Array.isArray(e)) return e;
}
function ho(e, t) {
	var n = e == null ? null : typeof Symbol < "u" && e[Symbol.iterator] || e["@@iterator"];
	if (n != null) {
		var r, i, a, o, s = [], c = !0, l = !1;
		try {
			if (a = (n = n.call(e)).next, t === 0) {
				if (Object(n) !== n) return;
				c = !1;
			} else for (; !(c = (r = a.call(n)).done) && (s.push(r.value), s.length !== t); c = !0);
		} catch (e) {
			l = !0, i = e;
		} finally {
			try {
				if (!c && n.return != null && (o = n.return(), Object(o) !== o)) return;
			} finally {
				if (l) throw i;
			}
		}
		return s;
	}
}
function go() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function _o(e, t) {
	return mo(e) || ho(e, t) || vo(e, t) || go();
}
function vo(e, t) {
	if (e) {
		if (typeof e == "string") return po(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? po(e, t) : void 0;
	}
}
function yo(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof fo;
			Promise.resolve(s ? o.v : o).then(function(n) {
				if (s) {
					var c = t === "return" && o.k ? t : "next";
					if (!o.k || n.done) return r(c, n);
					n = e[c](n).value;
				}
				i(!!a.done, n);
			}, function(e) {
				r("throw", e);
			});
		} catch (e) {
			i(2, e);
		}
	}
	function i(e, i) {
		e === 2 ? t.reject(i) : t.resolve({
			value: i,
			done: e
		}), (t = t.next) ? r(t.key, t.arg) : n = null;
	}
	this._invoke = function(e, i) {
		return new Promise(function(a, o) {
			var s = {
				key: e,
				arg: i,
				resolve: a,
				reject: o,
				next: null
			};
			n ? n = n.next = s : (t = n = s, r(e, i));
		});
	}, typeof e.return != "function" && (this.return = void 0);
}
yo.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, yo.prototype.next = function(e) {
	return this._invoke("next", e);
}, yo.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, yo.prototype.return = function(e) {
	return this._invoke("return", e);
};
var bo = Object.entries, xo = Object.setPrototypeOf, So = Object.isFrozen, Co = Object.getPrototypeOf, wo = Object.getOwnPropertyDescriptor, To = Object.freeze, Eo = Object.seal, Do = Object.create, Oo = typeof Reflect < "u" && Reflect, ko = Oo.apply, Ao = Oo.construct;
To ||= function(e) {
	return e;
}, Eo ||= function(e) {
	return e;
}, ko ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, Ao ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var jo = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var Mo = Z(Array.prototype.lastIndexOf), No = Z(Array.prototype.pop), Po = Z(Array.prototype.push), Fo = Z(Array.prototype.slice), Io = Z(Array.prototype.splice), Lo = Array.isArray, Ro = ts(Set), zo = Z(Set.prototype.add), Bo = Z(Set.prototype.has), Vo = Z(String.prototype.toLowerCase), Ho = Z(String.prototype.toString), Uo = Z(String.prototype.match), Wo = Z(String.prototype.replace), Go = Z(String.prototype.indexOf), Ko = Z(String.prototype.trim), qo = Z(Number.prototype.toString), Jo = Z(Boolean.prototype.toString), Yo = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), Xo = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), Zo = Z(Object.prototype.hasOwnProperty), Qo = Z(Object.prototype.toString), $o = Z(RegExp.prototype.test), es = ts(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return ko(e, t, n);
	};
}
function ts(e) {
	return function() {
		return Ao(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : Vo;
	if (xo && xo(e, null), !Lo(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (So(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function ns(e) {
	for (let t = 0; t < e.length; t++) Zo(e, t) || (e[t] = null);
	return e;
}
function rs(e) {
	let t = Do(null);
	for (let r of bo(e)) {
		var n = _o(r, 2);
		let i = n[0], a = n[1];
		Zo(e, i) && (t[i] = Lo(a) ? ns(a) : a && typeof a == "object" && a.constructor === Object ? rs(a) : a);
	}
	return t;
}
function is(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return qo(e);
		case "boolean": return Jo(e);
		case "bigint": return Yo ? Yo(e) : "0";
		case "symbol": return Xo ? Xo(e) : "Symbol()";
		case "undefined": return Qo(e);
		case "function":
		case "object": {
			if (e === null) return Qo(e);
			let t = e, n = as(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : Qo(e);
			}
			return Qo(e);
		}
		default: return Qo(e);
	}
}
function as(e, t) {
	for (; e !== null;) {
		let n = wo(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = Co(e);
	}
	function n() {
		return null;
	}
	return n;
}
function os(e, t) {
	for (; e !== null;) {
		let n = wo(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = Co(e);
	}
	return null;
}
function ss(e) {
	try {
		return $o(e, ""), !0;
	} catch {
		return !1;
	}
}
var cs = To(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), ls = To(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), us = To([
	"feBlend",
	"feColorMatrix",
	"feComponentTransfer",
	"feComposite",
	"feConvolveMatrix",
	"feDiffuseLighting",
	"feDisplacementMap",
	"feDistantLight",
	"feDropShadow",
	"feFlood",
	"feFuncA",
	"feFuncB",
	"feFuncG",
	"feFuncR",
	"feGaussianBlur",
	"feImage",
	"feMerge",
	"feMergeNode",
	"feMorphology",
	"feOffset",
	"fePointLight",
	"feSpecularLighting",
	"feSpotLight",
	"feTile",
	"feTurbulence"
]), ds = To([
	"animate",
	"color-profile",
	"cursor",
	"discard",
	"font-face",
	"font-face-format",
	"font-face-name",
	"font-face-src",
	"font-face-uri",
	"foreignobject",
	"hatch",
	"hatchpath",
	"mesh",
	"meshgradient",
	"meshpatch",
	"meshrow",
	"missing-glyph",
	"script",
	"set",
	"solidcolor",
	"unknown",
	"use"
]), fs = To(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), ps = To([
	"maction",
	"maligngroup",
	"malignmark",
	"mlongdiv",
	"mscarries",
	"mscarry",
	"msgroup",
	"mstack",
	"msline",
	"msrow",
	"semantics",
	"annotation",
	"annotation-xml",
	"mprescripts",
	"none"
]), ms = To(["#text"]), hs = To(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), gs = To(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), _s = To(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), vs = To([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), ys = Eo(/{{[\w\W]*|^[\w\W]*}}/g), bs = Eo(/<%[\w\W]*|^[\w\W]*%>/g), xs = Eo(/\${[\w\W]*/g), Ss = Eo(/^data-[\-\w.\u00B7-\uFFFF]+$/), Cs = Eo(/^aria-[\-\w]+$/), ws = Eo(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), Ts = Eo(/^(?:\w+script|data):/i), Es = Eo(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), Ds = Eo(/^html$/i), Os = Eo(/^[a-z][.\w]*(-[.\w]+)+$/i), ks = Eo(/<[/\w!]/g), As = Eo(/<[/\w]/g), js = Eo(/<\/no(script|embed|frames)/i), Ms = Eo(/\/>/i), $ = {
	element: 1,
	attribute: 2,
	text: 3,
	cdataSection: 4,
	entityReference: 5,
	entityNode: 6,
	processingInstruction: 7,
	comment: 8,
	document: 9,
	documentType: 10,
	documentFragment: 11,
	notation: 12
}, Ns = To(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), Ps = function() {
	return typeof window > "u" ? null : window;
}, Fs = function(e, t) {
	if (typeof e != "object" || typeof e.createPolicy != "function") return null;
	let n = null, r = "data-tt-policy-suffix";
	t && t.hasAttribute(r) && (n = t.getAttribute(r));
	let i = "dompurify" + (n ? "#" + n : "");
	try {
		return e.createPolicy(i, {
			createHTML(e) {
				return e;
			},
			createScriptURL(e) {
				return e;
			}
		});
	} catch {
		return console.warn("TrustedTypes policy " + i + " could not be created."), null;
	}
}, Is = function() {
	return {
		afterSanitizeAttributes: [],
		afterSanitizeElements: [],
		afterSanitizeShadowDOM: [],
		beforeSanitizeAttributes: [],
		beforeSanitizeElements: [],
		beforeSanitizeShadowDOM: [],
		uponSanitizeAttribute: [],
		uponSanitizeElement: [],
		uponSanitizeShadowNode: []
	};
}, Ls = function(e, t, n, r) {
	return Zo(e, t) && Lo(e[t]) ? Q(r.base ? rs(r.base) : {}, e[t], r.transform) : n;
}, Rs = function(e, t, n) {
	let r = Zo(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? rs(r) : n();
};
function zs() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : Ps(), t = (e) => zs(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = as(d, "cloneNode"), p = as(d, "remove"), m = as(d, "removeAttributeNode"), h = as(d, "nextSibling"), g = as(d, "childNodes"), _ = as(d, "parentNode"), v = as(d, "shadowRoot"), y = as(d, "attributes"), b = o && o.prototype ? as(o.prototype, "nodeType") : null, x = o && o.prototype ? as(o.prototype, "nodeName") : null, S = o && o.prototype ? as(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, ee = function(e) {
		return x ? x(e) : e.nodeName;
	}, te = os(d, "namespaceURI"), ne = function(e) {
		return te ? te(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let w, re = "", ie, ae = !1, oe = 0, se = function() {
		if (oe > 0) throw es("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
	}, ce = function(e) {
		se(), oe++;
		try {
			return w.createHTML(e);
		} finally {
			oe--;
		}
	}, le = function(e) {
		se(), oe++;
		try {
			return w.createScriptURL(e);
		} finally {
			oe--;
		}
	}, ue = function() {
		return ae ||= (ie = Fs(u, i), !0), ie;
	}, de = n, fe = de.implementation, pe = de.createNodeIterator, me = de.createDocumentFragment, he = de.getElementsByTagName, ge = r.importNode, T = Is();
	t.isSupported = typeof bo == "function" && typeof _ == "function" && fe && fe.createHTMLDocument !== void 0;
	let _e = ys, ve = bs, ye = xs, be = Ss, E = Cs, xe = Ts, D = Es, Se = Os, Ce = ws, O = null, we = Q({}, [
		...cs,
		...ls,
		...us,
		...fs,
		...ms
	]), k = null, Te = Q({}, [
		...hs,
		...gs,
		..._s,
		...vs
	]), Ee = Object.seal(Do(null, {
		tagNameCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		attributeNameCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		allowCustomizedBuiltInElements: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: !1
		}
	})), De = null, Oe = null, ke = Object.seal(Do(null, {
		tagCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		attributeCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		}
	})), Ae = !0, je = !0, Me = !1, Ne = !0, Pe = !1, Fe = !0, Ie = !1, Le = !1, Re = null, ze = null, Be = !1, Ve = !1, He = !1, Ue = !1, We = !0, Ge = !1, Ke = "user-content-", qe = !0, Je = !1, Ye = null, Xe = {}, A = null, Ze = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Qe = null, $e = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), et = null, tt = Q({}, [
		"alt",
		"class",
		"for",
		"id",
		"label",
		"name",
		"pattern",
		"placeholder",
		"role",
		"summary",
		"title",
		"value",
		"style",
		"xmlns"
	]), nt = "http://www.w3.org/1998/Math/MathML", rt = "http://www.w3.org/2000/svg", it = "http://www.w3.org/1999/xhtml", at = it, ot = !1, st = null, ct = Q({}, [
		nt,
		rt,
		it
	], Ho), lt = To([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), ut = Q({}, lt), dt = To(["annotation-xml"]), ft = Q({}, dt), pt = Q({}, [
		"title",
		"style",
		"font",
		"a",
		"script"
	]), mt = null, ht = ["application/xhtml+xml", "text/html"], j = null, gt = null, M = n.createElement("form"), _t = function(e) {
		return e instanceof RegExp || e instanceof Function;
	}, vt = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		if (gt && gt === e) return;
		e && typeof e == "object" || (e = {}), e = rs(e), mt = ht.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, j = mt === "application/xhtml+xml" ? Ho : Vo, O = Ls(e, "ALLOWED_TAGS", we, { transform: j }), k = Ls(e, "ALLOWED_ATTR", Te, { transform: j }), st = Ls(e, "ALLOWED_NAMESPACES", ct, { transform: Ho }), et = Ls(e, "ADD_URI_SAFE_ATTR", tt, {
			transform: j,
			base: tt
		}), Qe = Ls(e, "ADD_DATA_URI_TAGS", $e, {
			transform: j,
			base: $e
		}), A = Ls(e, "FORBID_CONTENTS", Ze, { transform: j }), De = Ls(e, "FORBID_TAGS", rs({}), { transform: j }), Oe = Ls(e, "FORBID_ATTR", rs({}), { transform: j }), Xe = Zo(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? rs(e.USE_PROFILES) : e.USE_PROFILES : !1, Ae = e.ALLOW_ARIA_ATTR !== !1, je = e.ALLOW_DATA_ATTR !== !1, Me = e.ALLOW_UNKNOWN_PROTOCOLS || !1, Ne = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Pe = e.SAFE_FOR_TEMPLATES || !1, Fe = e.SAFE_FOR_XML !== !1, Ie = e.WHOLE_DOCUMENT || !1, Ve = e.RETURN_DOM || !1, He = e.RETURN_DOM_FRAGMENT || !1, Ue = e.RETURN_TRUSTED_TYPE || !1, Be = e.FORCE_BODY || !1, We = e.SANITIZE_DOM !== !1, Ge = e.SANITIZE_NAMED_PROPS || !1, qe = e.KEEP_CONTENT !== !1, Je = e.IN_PLACE || !1, Ce = ss(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : ws, at = typeof e.NAMESPACE == "string" ? e.NAMESPACE : it, ut = Rs(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, lt)), ft = Rs(e, "HTML_INTEGRATION_POINTS", () => Q({}, dt));
		let t = Rs(e, "CUSTOM_ELEMENT_HANDLING", () => Do(null));
		if (Ee = Do(null), Zo(t, "tagNameCheck") && _t(t.tagNameCheck) && (Ee.tagNameCheck = t.tagNameCheck), Zo(t, "attributeNameCheck") && _t(t.attributeNameCheck) && (Ee.attributeNameCheck = t.attributeNameCheck), Zo(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (Ee.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), Eo(Ee), Pe && (je = !1), He && (Ve = !0), Xe && (O = Q({}, ms), k = Do(null), Xe.html === !0 && (Q(O, cs), Q(k, hs)), Xe.svg === !0 && (Q(O, ls), Q(k, gs), Q(k, vs)), Xe.svgFilters === !0 && (Q(O, us), Q(k, gs), Q(k, vs)), Xe.mathMl === !0 && (Q(O, fs), Q(k, _s), Q(k, vs))), ke.tagCheck = null, ke.attributeCheck = null, Zo(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? ke.tagCheck = e.ADD_TAGS : Lo(e.ADD_TAGS) && (O === we && (O = rs(O)), Q(O, e.ADD_TAGS, j))), Zo(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? ke.attributeCheck = e.ADD_ATTR : Lo(e.ADD_ATTR) && (k === Te && (k = rs(k)), Q(k, e.ADD_ATTR, j))), Zo(e, "ADD_FORBID_CONTENTS") && Lo(e.ADD_FORBID_CONTENTS) && (A === Ze && (A = rs(A)), Q(A, e.ADD_FORBID_CONTENTS, j)), qe && (O["#text"] = !0), Ie && Q(O, [
			"html",
			"head",
			"body"
		]), O.table && (Q(O, ["tbody"]), delete De.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw es("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw es("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = w;
			w = e.TRUSTED_TYPES_POLICY;
			try {
				re = ce("");
			} catch (e) {
				throw w = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (w = void 0, re = "") : (w === void 0 && (w = ue()), w && typeof re == "string" && (re = ce("")));
		To && To(e), gt = e;
	}, yt = Q({}, [
		...ls,
		...us,
		...ds
	]), bt = Q({}, [...fs, ...ps]), xt = function(e, t, n) {
		return t.namespaceURI === it ? e === "svg" : t.namespaceURI === nt ? e === "svg" && (n === "annotation-xml" || ut[n]) : !!yt[e];
	}, St = function(e, t, n) {
		return t.namespaceURI === it ? e === "math" : t.namespaceURI === rt ? e === "math" && ft[n] : !!bt[e];
	}, Ct = function(e, t, n) {
		return t.namespaceURI === rt && !ft[n] || t.namespaceURI === nt && !ut[n] ? !1 : !bt[e] && (pt[e] || !yt[e]);
	}, wt = function(e) {
		let t = ne(e), n = Vo(ee(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: ne(r) }, a = Vo(ee(r))) : (i = { namespaceURI: at }, a = "template"), st[t] ? t === rt ? xt(n, i, a) : t === nt ? St(n, i, a) : t === it ? Ct(n, i, a) : !!(mt === "application/xhtml+xml" && st[t]) : !1;
	}, Tt = function(e) {
		Po(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw es("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, Et = function(e, t, n) {
		try {
			m(e, t);
		} catch {
			try {
				e.removeAttribute(n);
			} catch {}
		}
	}, Dt = function(e) {
		jt(e);
		let t = g(e);
		t && jo(Fo(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === $.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && Et(e, r, i);
		}
	}, Ot = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		Po(t.removed, {
			attribute: r || null,
			from: n
		});
		try {
			r ? m(n, r) : n.removeAttribute(e);
		} catch {
			try {
				n.removeAttribute(e);
			} catch {}
		}
		if (e === "is") {
			if (Ve || He) try {
				Tt(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, kt = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || k[j(i)] || Et(e, r, i);
		}
	}, At = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === $.element) {
				let t = v(e);
				Wt(t) && r.push(t);
			}
			let i = g(e);
			if (i) for (let e = i.length - 1; e >= 0; --e) r.push(i[e]);
		}
	}, jt = function(e) {
		At(e, (e) => {
			C(e) === $.element && kt(e);
		});
	}, Mt = function(e) {
		let t = [];
		return At(e, (e) => {
			t.push(e);
		}, !0), t;
	}, Nt = function(e, t) {
		let n = Ro();
		jo(Mt(e), (e) => {
			C(e) === $.element && zo(n, e);
		}), jo(t, (e) => {
			C(e) === $.element && !Bo(n, e) && kt(e);
		});
	}, Pt = function(e, n) {
		let r = !1;
		return jo(t.removed, (t) => {
			t.element && (r ||= t.element === e, jt(t.element));
		}), Nt(e, n), r;
	}, Ft = function(e, t) {
		let n = Ye, r = null, i = () => (r || (r = Ro(), jo(t, (e) => {
			zo(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = Mt(e), a = [e], o = !1;
			if (jo(r, (e) => {
				C(e) === $.documentFragment ? Po(a, e) : !o && !Bo(n, e) && (o = Bo(i(), e));
			}), !o) return;
			if (t >= 8) throw es("the tree kept changing during in-place sanitization; refusing to return it");
			jo(a, (t, n) => {
				It(t, n > 0, e, i());
			});
		}
	}, It = function(e, t, n, r) {
		let i = zt(e), a;
		for (; a = i.nextNode();) Bo(r, a) && !Bo(Ye, a) && (t && Gt(T.uponSanitizeShadowNode, a, null), Xt(a, n), en(a, n), Wt(a.content) && tn(a.content));
	}, Lt = function(e) {
		Ye && zo(Ye, e);
	}, N = function(e, t) {
		return Fe ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, Rt = function(e) {
		Fe && At(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && $o(As, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = j(ee(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && N("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, P = function(e) {
		let t = null, r = null;
		if (Be) e = "<remove></remove>" + e;
		else {
			let t = Uo(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		mt === "application/xhtml+xml" && at === it && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = w ? ce(e) : e;
		if (at === it) try {
			t = new l().parseFromString(i, mt);
		} catch {}
		if (!t || !t.documentElement) {
			t = fe.createDocument(at, "template", null);
			try {
				t.documentElement.innerHTML = ot ? re : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), at === it ? he.call(t, Ie ? "html" : "body")[0] : Ie ? t.documentElement : a;
	}, zt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return pe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, Bt = function(e) {
		return e = Wo(e, _e, " "), e = Wo(e, ve, " "), e = Wo(e, ye, " "), e;
	}, Vt = function(e) {
		Ht(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && jo(t, (e) => {
			Wt(e.content) && Vt(e.content);
		});
	}, Ht = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = pe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = Bt(r.data), r = n.nextNode();
	}, Ut = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || j(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, Wt = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return b(e) === $.documentFragment;
		} catch {
			return !1;
		}
	}, F = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return typeof b(e) == "number";
		} catch {
			return !1;
		}
	};
	function Gt(e, n, r) {
		e.length !== 0 && jo(e, (e) => {
			e.call(t, n, r, gt);
		});
	}
	let Kt = function(e, t) {
		return !!(Fe && e.hasChildNodes() && !F(e.firstElementChild) && $o(ks, e.textContent) && $o(ks, e.innerHTML) || Fe && Ns[t] && C(e) === $.element && ne(e) === it && (F(e.firstElementChild) || typeof e.textContent == "string" && $o(ks, e.textContent)) || e.nodeType === $.processingInstruction || Fe && e.nodeType === $.comment && $o(As, e.data));
	}, qt = function(e, t) {
		return e instanceof RegExp ? $o(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, Jt = function(e, t, n) {
		if (!De[t] && Qt(t) && qt(Ee.tagNameCheck, t)) return !1;
		if (qe && !A[t] && !A[Vo(t)]) {
			let t = _(e), r = g(e);
			if (r && t) {
				let i = r.length;
				for (let a = i - 1; a >= 0; --a) {
					let i = e === n ? f(r[a], !0) : r[a];
					t.insertBefore(i, h(e));
				}
			}
		}
		return Tt(e), !0;
	}, Yt = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? rs(t) : t;
	}, I = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Je && jt(e), !0);
	}, Xt = function(e, n) {
		if (Lt(e), Gt(T.beforeSanitizeElements, e, null), I(e, n)) return !0;
		if (Ut(e)) return Tt(e), !0;
		let r = j(ee(e));
		if (O = Yt(T.uponSanitizeElement, O, we, Re), Gt(T.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: O
		}), I(e, n)) return !0;
		if (Kt(e, r)) return Tt(e), !0;
		if (De[r] || !(ke.tagCheck instanceof Function && ke.tagCheck(r)) && !O[r]) {
			let t = Jt(e, r, n);
			return t === !1 && (Gt(T.afterSanitizeElements, e, null), I(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !wt(e) || (r === "noscript" || r === "noembed" || r === "noframes") && $o(js, e.innerHTML)) return Tt(e), !0;
		if (Pe && e.nodeType === $.text) {
			let n = Bt(e.textContent);
			e.textContent !== n && (Po(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return Gt(T.afterSanitizeElements, e, null), I(e, n);
	}, Zt = function(e, t, r) {
		if (Oe[t] || N(t, e) || We && (t === "id" || t === "name") && (r in n || r in M)) return !1;
		let i = k[t] || ke.attributeCheck instanceof Function && ke.attributeCheck(t, e);
		return je && $o(be, t) || Ae && $o(E, t) ? !0 : i ? et[t] || $o(Ce, Wo(r, D, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && Go(r, "data:") === 0 && Qe[e] || Me && !$o(xe, Wo(r, D, "")) ? !0 : !r : Qt(e) && qt(Ee.tagNameCheck, e) && qt(Ee.attributeNameCheck, t, e) || t === "is" && Ee.allowCustomizedBuiltInElements && qt(Ee.tagNameCheck, r);
	}, L = Q({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), Qt = function(e) {
		return !L[Vo(e)] && $o(Se, e);
	}, R = function(e, t, n, r) {
		if (w && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return ce(r);
			case "TrustedScriptURL": return le(r);
		}
		return r;
	}, z = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !Ut(e) || (Tt(e), !1);
		} catch {
			return Ot(t, e), !1;
		}
	}, $t = function(e, t) {
		return Fe && $o(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : Vo(e) === "attributename" && Uo(t, "href") !== null;
	}, en = function(e, n) {
		if (Gt(T.beforeSanitizeAttributes, e, null), I(e, n)) return;
		let r = e.attributes;
		if (!r || Ut(e)) return;
		k = Yt(T.uponSanitizeAttribute, k, Te, ze);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: k,
			forceKeepAttr: void 0
		}, a = r.length, o = j(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = j(s), d = l, f = s === "value" ? d : Ko(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, Gt(T.uponSanitizeAttribute, e, i), f = i.attrValue, Ge && (u === "id" || u === "name") && Go(f, Ke) !== 0 && (Ot(s, e, n), f = Ke + f, p = !0), $t(s, f) ? Ot(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !Ne && $o(Ms, f) ? Ot(s, e, n) : (Pe && (f = Bt(f)), Zt(o, u, f) ? (f = R(o, u, c, f), f !== d && z(e, s, c, f) && p && No(t.removed)) : Ot(s, e, n)) : Ot(s, e, n));
		}
		Gt(T.afterSanitizeAttributes, e, null), I(e, n);
	}, tn = function(e) {
		let t = null, n = zt(e);
		for (Gt(T.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (Gt(T.uponSanitizeShadowNode, t, null), Xt(t, e), en(t, e), Wt(t.content) && tn(t.content), C(t) === $.element) {
			let e = v(t);
			Wt(e) && (nn(e), tn(e));
		}
		Pe && Ht(e), Gt(T.afterSanitizeShadowDOM, e, null);
	}, nn = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				tn(e.shadow);
				continue;
			}
			let n = e.node, r = C(n) === $.element, i = g(n);
			if (i) for (let e = i.length - 1; e >= 0; --e) t.push({
				node: i[e],
				shadow: null
			});
			if (r) {
				let e = x ? x(n) : null;
				if (typeof e == "string" && j(e) === "template") {
					let e = n.content;
					Wt(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				Wt(e) && t.push({
					node: null,
					shadow: e
				}, {
					node: e,
					shadow: null
				});
			}
		}
	};
	return t.sanitize = function(e) {
		let n = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {}, i = null, a = null, o = null, s = null;
		if (ot = !e, ot && (e = "<!-->"), typeof e != "string" && !F(e) && (e = is(e), typeof e != "string")) throw es("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Le ? (O = Re, k = ze) : vt(n), (T.uponSanitizeElement.length > 0 || T.uponSanitizeAttribute.length > 0) && (O = rs(O)), T.uponSanitizeAttribute.length > 0 && (k = rs(k)), t.removed = [];
		let c = Je && typeof e != "string" && F(e), l = [], u = Ye;
		if (c) {
			Rt(e);
			let t = ee(e);
			if (typeof t == "string") {
				let n = j(t);
				if (!O[n] || De[n]) throw Dt(e), es("root node is forbidden and cannot be sanitized in-place");
			}
			if (Ut(e)) throw Dt(e), es("root node is clobbered and cannot be sanitized in-place");
			l = Mt(e), Ye = Ro();
		} else if (F(e)) i = P("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), nn(i);
		else {
			if (!Ve && !Pe && !Ie && e.indexOf("<") === -1) return w && Ue ? ce(e) : e;
			if (i = P(e), !i) return Ve ? null : Ue ? re : "";
		}
		i && Be && Tt(i.firstChild);
		let d = c ? e : i;
		try {
			c && nn(e);
			let t = zt(d);
			for (; o = t.nextNode();) Xt(o, d), en(o, d), Wt(o.content) && tn(o.content);
			c && Ft(e, l);
		} catch (t) {
			throw c && (Dt(e), Pt(e, l)), t;
		} finally {
			Ye = u;
		}
		if (c) {
			if (Pt(e, l)) throw es("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Pe && Vt(e), e;
		}
		if (Ve) {
			if (Pe && Vt(i), He) for (s = me.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (k.shadowroot || k.shadowrootmode) && (s = ge.call(r, s, !0)), s;
		}
		let f = Ie ? i.outerHTML : i.innerHTML;
		return Ie && O["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && $o(Ds, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Pe && (f = Bt(f)), w && Ue ? ce(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		vt(e), Le = !0, Re = O, ze = k;
	}, t.clearConfig = function() {
		gt = null, Le = !1, Re = null, ze = null, w = ie, re = "";
	}, t.isValidAttribute = function(e, t, n) {
		gt || vt({});
		let r = j(e), i = j(t);
		return typeof n == "string" && ($t(t, n) || !Ne && $o(Ms, n)) ? !1 : Zt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && Zo(T, e) && Po(T[e], t);
	}, t.removeHook = function(e, t) {
		if (Zo(T, e)) {
			if (t !== void 0) {
				let n = Mo(T[e], t);
				return n === -1 ? void 0 : Io(T[e], n, 1)[0];
			}
			return No(T[e]);
		}
	}, t.removeHooks = function(e) {
		Zo(T, e) && (T[e] = []);
	}, t.removeAllHooks = function() {
		T = Is();
	}, t;
}
var Bs = zs();
// kit/markdown.js
function Vs(e) {
	let t = X.parse(e || "", {
		async: !1,
		gfm: !0,
		breaks: !1
	}), n = Bs.sanitize(t, {
		USE_PROFILES: { html: !0 },
		FORBID_TAGS: [
			"style",
			"form",
			"input",
			"button",
			"iframe",
			"object",
			"embed",
			"svg",
			"math",
			"img"
		],
		FORBID_ATTR: ["style"]
	}), r = new DOMParser().parseFromString(n, "text/html");
	for (let e of r.querySelectorAll("a")) {
		let t = !1;
		try {
			t = ["http:", "https:"].includes(new URL(e.href).protocol);
		} catch {}
		t || e.removeAttribute("href"), e.target = "_blank", e.rel = "noopener noreferrer";
	}
	return r.body.innerHTML;
}
// kit/Holder.svelte
var Hs = /* @__PURE__ */ W("<span class=\"svelte-bwb1ob\"> </span>"), Us = /* @__PURE__ */ W("<span class=\"name svelte-bwb1ob\" data-testid=\"holder-pool\"> </span>"), Ws = /* @__PURE__ */ W("<span class=\"name svelte-bwb1ob\" data-testid=\"holder-name\"> </span>"), Gs = /* @__PURE__ */ W("<span class=\"faint model svelte-bwb1ob\"> </span>"), Ks = /* @__PURE__ */ W("<span class=\"tool muted svelte-bwb1ob\" data-testid=\"holder-tool\"> </span>"), qs = /* @__PURE__ */ W("<button type=\"button\" class=\"holder svelte-bwb1ob\" data-testid=\"holder\"><span></span> <!> <!> <!> <!></button>");
function Js(e, t) {
	He(t, !0);
	let n = ri(t, "label", 3, ""), r = /* @__PURE__ */ ut(() => t.h.display_name || t.h.id), i = /* @__PURE__ */ ut(() => t.h.tool || t.h.tool_target ? pi(t.h.agent, {
		title: t.h.tool,
		target: t.h.tool_target,
		intent: t.h.tool_intent
	}).label : "working"), a = /* @__PURE__ */ ut(() => [
		U(r),
		[t.h.pool, t.h.model].filter(Boolean).join(" · "),
		t.h.busy ? U(i) : "",
		t.h.id
	].filter(Boolean).join("\n") + "\nClick to open read-only");
	var o = qs(), s = L(o);
	let c;
	var l = z(s, 2), u = (e) => {
		var t = Hs(), r = R(t, !0);
		B(() => K(r, n())), G(e, t);
	};
	q(l, (e) => {
		n() && e(u);
	});
	var d = z(l, 2), f = (e) => {
		var n = Us(), r = R(n, !0);
		B(() => K(r, t.h.pool)), G(e, n);
	}, p = (e) => {
		var t = Ws(), n = R(t, !0);
		B(() => K(n, U(r))), G(e, t);
	};
	q(d, (e) => {
		t.h.pool ? e(f) : e(p, -1);
	});
	var m = z(d, 2), h = (e) => {
		var n = Gs(), r = R(n, !0);
		B(() => K(r, t.h.model)), G(e, n);
	};
	q(m, (e) => {
		t.h.model && e(h);
	});
	var g = z(m, 2), _ = (e) => {
		var t = Ks(), n = R(t, !0);
		B(() => K(n, U(i))), G(e, t);
	};
	q(g, (e) => {
		t.h.busy && e(_);
	}), O(o), B(() => {
		Jr(o, "title", U(a)), Jr(o, "data-session", t.h.id), Jr(o, "data-name", U(r)), c = Ur(s, 1, "dot svelte-bwb1ob", null, c, {
			busy: t.h.busy,
			dead: t.h.alive === !1
		});
	}), cr("click", o, (e) => {
		e.stopPropagation(), t.onopen(t.h.id);
	}), G(e, o), Ue();
}
lr(["click"]);
// plugins/attention/Attention.svelte
var Ys = /* @__PURE__ */ W("<button class=\"ghost\" data-testid=\"push-test\">Test</button>"), Xs = /* @__PURE__ */ W("<!> <button data-testid=\"push-toggle\"> </button>", 1), Zs = /* @__PURE__ */ W("<span class=\"faint pushna svelte-t4vask\" data-testid=\"push-state\"> </span>"), Qs = /* @__PURE__ */ W("<div class=\"card err\"> </div>"), $s = /* @__PURE__ */ W("<p class=\"faint\">Loading…</p>"), ec = /* @__PURE__ */ W("<div class=\"card empty\" data-testid=\"attention-empty\"><div class=\"big\">Nothing needs you.</div> <div class=\"muted\">No escalations, questions, decisions or other items waiting on you.</div></div>"), tc = /* @__PURE__ */ W("<a target=\"_blank\" rel=\"noopener noreferrer\" data-testid=\"attention-ref\"> </a>"), nc = /* @__PURE__ */ W("· <!>", 1), rc = /* @__PURE__ */ W("<div class=\"summary muted svelte-t4vask\" data-testid=\"attention-summary\"> </div>"), ic = /* @__PURE__ */ W("<details class=\"qctx svelte-t4vask\" data-testid=\"attention-context\"><summary class=\"svelte-t4vask\">Context</summary> <div class=\"md svelte-t4vask\"></div></details>"), ac = /* @__PURE__ */ W("<span class=\"muted\" data-testid=\"row-pending\"> </span>"), oc = /* @__PURE__ */ W("<span class=\"warn-text svelte-t4vask\" data-testid=\"row-stuck\"> </span>"), sc = /* @__PURE__ */ W("<span class=\"rec svelte-t4vask\"></span>"), cc = /* @__PURE__ */ W("<span class=\"od svelte-t4vask\"> </span>"), lc = /* @__PURE__ */ W("<button type=\"button\" data-testid=\"question-option\"><span class=\"ot svelte-t4vask\"> <!></span> <!></button>"), uc = /* @__PURE__ */ W("<div class=\"opts svelte-t4vask\" data-testid=\"question-options\"></div>"), dc = /* @__PURE__ */ W("<form class=\"qans svelte-t4vask\" data-testid=\"question-answer\"><input type=\"text\" placeholder=\"Answer…\" aria-label=\"Answer\" data-testid=\"question-answer-text\" class=\"svelte-t4vask\"/> <button type=\"submit\" class=\"primary svelte-t4vask\" data-testid=\"question-answer-send\">Answer</button></form>"), fc = /* @__PURE__ */ W("<button type=\"button\" data-testid=\"row-resolve\" class=\"svelte-t4vask\"> </button>"), pc = /* @__PURE__ */ W("<button type=\"button\" class=\"ghost svelte-t4vask\" data-testid=\"open-chat\" title=\"Talk it over or answer in your own words\">Discuss</button>"), mc = /* @__PURE__ */ W("<button type=\"button\" class=\"ghost svelte-t4vask\" data-testid=\"open-chat\">Open chat</button>"), hc = /* @__PURE__ */ W("<!> <!> <!> <!>", 1), gc = /* @__PURE__ */ W("<tr data-testid=\"attention-row\"><td class=\"c-id svelte-t4vask\"><span class=\"pill warn\"> </span></td><td class=\"c-title svelte-t4vask\"> <div class=\"faint mono\"> <!></div> <!> <div class=\"acts svelte-t4vask\" role=\"presentation\"><!> <!></div></td><td class=\"muted full svelte-t4vask\"> </td><td class=\"muted c-time svelte-t4vask\"> </td></tr>"), _c = /* @__PURE__ */ W("<div class=\"card scroll\"><table class=\"stack\" data-testid=\"attention-list\"><thead><tr><th>Kind</th><th>Item</th><th>Why</th><th>Since</th></tr></thead><tbody></tbody></table></div>"), vc = /* @__PURE__ */ W("<span class=\"step muted svelte-t4vask\"> </span> <!>", 1), yc = /* @__PURE__ */ W("<span class=\"faint\" data-testid=\"no-holder\">waiting for a session</span>"), bc = /* @__PURE__ */ W("<a class=\"mono svelte-t4vask\"> </a>"), xc = /* @__PURE__ */ W("<span class=\"note svelte-t4vask\"> </span>"), Sc = /* @__PURE__ */ W("<li class=\"svelte-t4vask\"><span> </span> <span class=\"step muted svelte-t4vask\"> </span> <!> <!> <span class=\"faint when svelte-t4vask\"> </span></li>"), Cc = /* @__PURE__ */ W("<li class=\"svelte-t4vask\"><span class=\"pill\"> </span><span class=\"step muted svelte-t4vask\"> </span><span class=\"faint\">no report yet</span></li>"), wc = /* @__PURE__ */ W("<li class=\"svelte-t4vask\"><span class=\"faint\">waiting for a step to start</span></li>"), Tc = /* @__PURE__ */ W("<a class=\"svelte-t4vask\"> </a>"), Ec = /* @__PURE__ */ W("<li class=\"svelte-t4vask\"><span class=\"pill\">comment</span><span class=\"step muted mono svelte-t4vask\"><!></span><span class=\"note svelte-t4vask\"> </span><span class=\"faint when svelte-t4vask\"> </span></li>"), Dc = /* @__PURE__ */ W("<div class=\"item card svelte-t4vask\" data-testid=\"progress-card\"><button class=\"head svelte-t4vask\"><span class=\"dot accent\"></span> <span class=\"title svelte-t4vask\"> </span> <span class=\"mono faint id svelte-t4vask\"> </span> <span class=\"faint when svelte-t4vask\"> </span></button> <div class=\"who svelte-t4vask\"><!></div> <ul class=\"svelte-t4vask\"><!> <!></ul></div>"), Oc = /* @__PURE__ */ W("<div class=\"progress svelte-t4vask\" data-testid=\"progress\"></div>"), kc = /* @__PURE__ */ W("<div class=\"faint\">Loading…</div>"), Ac = /* @__PURE__ */ W("<div class=\"faint none svelte-t4vask\" data-testid=\"progress-empty\">No runs in progress.</div>"), jc = /* @__PURE__ */ W("<header class=\"vh\"><h1>Attention</h1> <span class=\"sub\">What waits on a human, oldest first.</span> <span class=\"actions\"><!></span></header> <div class=\"pad\"><!> <!> <h2 class=\"st\">In progress</h2> <!></div>", 1);
function Mc(e, t) {
	He(t, !0);
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
				e && (a.has(t) ? ++e.loads >= 2 ? (u[t] = e.label, delete l[t], ui(t)) : o = !0 : (delete l[t], ui(t)));
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
	dn(() => nr(() => {
		for (let [e, t] of ci()) p(e, t);
		let e = di((e, t) => {
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
			await ai(t.api, n) && li(r, n.label);
		} finally {
			P(h, "");
		}
	}
	let _ = (e, n) => ({
		label: "Answered " + n.title,
		run: () => t.api.call("question.answer", {
			id: e.id,
			option: n.id
		}),
		ask: "Answer " + e.id + " with option " + n.id + ": " + n.title + "?"
	}), v = F({}), y = (e, n) => ({
		label: "Answered",
		run: () => t.api.call("question.answer", {
			id: e.id,
			text: n
		}),
		ask: "Answer " + e.id + ": \"" + n + "\"?"
	});
	async function b(e) {
		let t = e.kind + ":" + e.id, n = (v[t] || "").trim();
		n && (await g(e, y(e, n)), l[t] && delete v[t]);
	}
	let x = {
		unsupported: "This browser has no Web Push (on iOS, add Desk to the home screen first).",
		unavailable: "Desk runs without --push.",
		denied: "Notifications are blocked in the browser settings.",
		off: "Get a notification for each new attention item.",
		on: "Notifications are on for this browser."
	};
	async function S() {
		P(c, U(c).state === "on" ? await t.api.push.unsubscribe() : await t.api.push.subscribe(), !0);
	}
	async function C() {
		P(c, await t.api.push.test(), !0), U(c).error || t.api.toast("Test notification sent");
	}
	function ee(e) {
		if (!e || e.startsWith("0001")) return "";
		let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3);
		return t < 60 ? t + "s ago" : t < 3600 ? Math.round(t / 60) + "m ago" : t < 86400 ? Math.round(t / 3600) + "h ago" : Math.round(t / 86400) + "d ago";
	}
	let te = (e, t) => typeof e?.[t] == "string" ? JSON.parse(e[t] || "{}") : e?.[t] || {};
	function ne(e) {
		return e.map((e) => ({
			s: e,
			r: te(e, "meta").report
		})).filter((e) => e.r && (e.r.note || e.r.state)).sort((e, t) => String(t.r.at).localeCompare(String(e.r.at)));
	}
	let w = (e) => e === "failed" || e === "blocked" ? "bad" : e === "waiting" || e === "dropped" ? "warn" : e === "ready" ? "ok" : "accent", re = (e) => !!e && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(e), ie = (e) => e?.status === "in_progress" && e.holder ? U(a)[e.holder] ?? { id: e.holder } : null, ae = (e) => t.api.openSession(e, { readonly: !0 });
	var oe = jc(), se = Qt(oe), ce = z(L(se), 4), le = L(ce), ue = (e) => {
		var t = Xs(), n = Qt(t), r = (e) => {
			var t = Ys();
			cr("click", t, C), G(e, t);
		};
		q(n, (e) => {
			U(c).state === "on" && e(r);
		});
		var i = z(n, 2);
		let a;
		var o = R(i, !0);
		B(() => {
			Jr(i, "title", x[U(c).state]), a = Ur(i, 1, "", null, a, { on: U(c).state === "on" }), K(o, U(c).state === "on" ? "Notifications on" : "Notify me");
		}), cr("click", i, S), G(e, t);
	}, de = (e) => {
		var t = Zs(), n = R(t, !0);
		B(() => {
			Jr(t, "title", x[U(c).state]), K(n, U(c).state === "unavailable" ? "" : "no notifications");
		}), G(e, t);
	};
	q(le, (e) => {
		U(c).state === "on" || U(c).state === "off" ? e(ue) : e(de, -1);
	}), O(ce), O(se);
	var fe = z(se, 2), pe = L(fe), me = (e) => {
		var t = Qs(), n = R(t, !0);
		B(() => K(n, U(c).error)), G(e, t);
	};
	q(pe, (e) => {
		U(c).error && e(me);
	});
	var he = z(pe, 2), ge = (e) => {
		var t = Qs(), n = R(t, !0);
		B(() => K(n, U(r))), G(e, t);
	}, T = (e) => {
		G(e, $s());
	}, _e = (e) => {
		G(e, ec());
	}, ve = (e) => {
		var r = _c(), i = L(r), a = z(L(i));
		Pr(a, 21, () => U(n), (e) => e.kind + e.id, (e, n) => {
			let r = /* @__PURE__ */ ut(() => U(n).kind + ":" + U(n).id), i = /* @__PURE__ */ ut(() => l[U(r)]);
			var a = gc();
			let o;
			var c = L(a), d = R(L(c), !0);
			O(c);
			var f = z(c), p = L(f, !0), y = z(p), x = L(y, !0), S = z(x), C = (e) => {
				var t = nc(), r = z(Qt(t)), i = (e) => {
					var t = tc(), r = R(t, !0);
					B(() => {
						Jr(t, "href", U(n).url), K(r, U(n).ref);
					}), cr("click", t, (e) => e.stopPropagation()), G(e, t);
				}, a = /* @__PURE__ */ ut(() => U(n).url && fi(U(n).url)), o = (e) => {
					var t = _r();
					B(() => K(t, U(n).ref)), G(e, t);
				};
				q(r, (e) => {
					U(a) ? e(i) : e(o, -1);
				}), G(e, t);
			};
			q(S, (e) => {
				U(n).ref && e(C);
			}), O(y);
			var te = z(y, 2), ne = (e) => {
				var t = rc(), r = R(t, !0);
				B(() => K(r, U(n).summary)), G(e, t);
			};
			q(te, (e) => {
				U(n).summary && e(ne);
			});
			var w = z(te, 2), re = L(w), ie = (e) => {
				var t = ic(), r = z(L(t), 2);
				Br(r, () => Vs(U(n).context), !0), O(r), O(t), G(e, t);
			};
			q(re, (e) => {
				U(n).context && e(ie);
			});
			var ae = z(re, 2), oe = (e) => {
				var t = ac(), n = R(t);
				B(() => K(n, `${U(i).label ?? ""}…`)), G(e, t);
			}, se = (e) => {
				var i = hc(), a = Qt(i), o = (e) => {
					var t = oc(), n = R(t);
					B(() => K(n, `still listed after ${u[U(r)] ?? ""}`)), G(e, t);
				};
				q(a, (e) => {
					u[U(r)] && e(o);
				});
				var c = z(a, 2), l = (e) => {
					var t = uc();
					Pr(t, 21, () => U(n).options, (e) => e.id, (e, t) => {
						var i = lc();
						let a;
						var o = L(i), s = L(o), c = z(s), l = (e) => {
							var t = sc();
							t.textContent = " (recommended)", G(e, t);
						};
						q(c, (e) => {
							U(t).recommended && e(l);
						}), O(o);
						var u = z(o, 2), d = (e) => {
							var n = cc(), r = R(n, !0);
							B(() => K(r, U(t).description)), G(e, n);
						};
						q(u, (e) => {
							U(t).description && e(d);
						}), O(i), B(() => {
							a = Ur(i, 1, "opt svelte-t4vask", null, a, { primary: U(t).recommended }), i.disabled = U(h) === U(r), Jr(i, "data-option", U(t).id), K(s, `${U(t).id ?? ""}. ${U(t).title ?? ""}`);
						}), cr("click", i, () => g(U(n), _(U(n), U(t)))), G(e, i);
					}), O(t), G(e, t);
				}, d = (e) => {
					var t = dc(), i = L(t);
					qr(i);
					var a = z(i, 2);
					O(t), B((e) => {
						i.disabled = U(h) === U(r), a.disabled = e;
					}, [() => U(h) === U(r) || !(v[U(r)] || "").trim()]), sr("submit", t, (e) => {
						e.preventDefault(), b(U(n));
					}), Qr(i, () => v[U(r)], (e) => v[U(r)] = e), G(e, t);
				};
				q(c, (e) => {
					U(n).options?.length ? e(l) : U(n).kind === "question" && e(d, 1);
				});
				var f = z(c, 2);
				Pr(f, 17, () => ii(t.api, U(n).kind, U(n).id), (e) => e.label, (e, t) => {
					var i = fc(), a = R(i, !0);
					B(() => {
						i.disabled = U(h) === U(r), K(a, U(t).label);
					}), cr("click", i, () => g(U(n), U(t))), G(e, i);
				});
				var p = z(f, 2), y = (e) => {
					var t = pc();
					B(() => t.disabled = U(s) === U(r)), cr("click", t, () => m(U(n))), G(e, t);
				}, x = (e) => {
					var t = mc();
					B(() => t.disabled = U(s) === U(r)), cr("click", t, () => m(U(n))), G(e, t);
				};
				q(p, (e) => {
					U(n).kind === "question" ? e(y) : e(x, -1);
				}), G(e, i);
			};
			q(ae, (e) => {
				U(i) ? e(oe) : e(se, -1);
			}), O(w), O(f);
			var ce = z(f), le = R(ce, !0), ue = R(z(ce), !0);
			O(a), B((e) => {
				o = Ur(a, 1, "click svelte-t4vask", null, o, {
					busy: U(s) === U(r) || !!U(i),
					pending: !!U(i)
				}), Jr(a, "data-item", U(n).kind + ":" + U(n).id), K(d, U(n).kind), K(p, U(n).title || U(n).id), K(x, U(n).id), K(le, U(n).why), K(ue, e);
			}, [() => ee(U(n).since)]), cr("click", a, () => m(U(n))), cr("click", w, (e) => e.stopPropagation()), G(e, a);
		}), O(a), O(i), O(r), G(e, r);
	};
	q(he, (e) => {
		U(r) ? e(ge) : U(n) ? U(n).length === 0 ? e(_e, 2) : e(ve, -1) : e(T, 1);
	});
	var ye = z(he, 4), be = (e) => {
		var n = Oc();
		Pr(n, 21, () => U(i), ({ run: e, steps: t, comment: n }) => e.id, (e, n) => {
			let r = () => U(n).run, i = () => U(n).steps, a = () => U(n).comment, o = /* @__PURE__ */ ut(() => ne(i())), s = /* @__PURE__ */ ut(() => i().find((e) => e.status === "in_progress")), c = /* @__PURE__ */ ut(() => ie(U(s)));
			var l = Dc(), u = L(l), d = z(L(u), 2), f = R(d, !0), p = z(d, 2), m = R(p, !0), h = R(z(p, 2), !0);
			O(u);
			var g = z(u, 2), _ = L(g), v = (e) => {
				var t = vc(), n = Qt(t), r = R(n, !0);
				Js(z(n, 2), {
					get h() {
						return U(c);
					},
					onopen: ae
				}), B((e) => K(r, e), [() => te(U(s), "fields").step || U(s).title]), G(e, t);
			}, y = (e) => {
				G(e, yc());
			};
			q(_, (e) => {
				U(c) ? e(v) : e(y, -1);
			}), O(g);
			var b = z(g, 2), x = L(b);
			Pr(x, 17, () => U(o).slice(0, 3), ({ s: e, r: t }) => e.id, (e, t) => {
				let n = () => U(t).s, r = () => U(t).r, i = /* @__PURE__ */ ut(() => ie(n()));
				var a = Sc(), o = L(a), s = R(o, !0), c = z(o, 2), l = R(c, !0), u = z(c, 2), d = (e) => {
					Js(e, {
						get h() {
							return U(i);
						},
						onopen: ae
					});
				}, f = (e) => {
					var t = bc(), n = R(t, !0);
					B(() => {
						Jr(t, "href", "?session=" + r().by), K(n, r().by);
					}), cr("click", t, (e) => {
						e.preventDefault(), ae(r().by);
					}), G(e, t);
				}, p = /* @__PURE__ */ ut(() => re(r().by));
				q(u, (e) => {
					U(i) ? e(d) : U(p) && e(f, 1);
				});
				var m = z(u, 2), h = (e) => {
					var t = xc(), n = R(t, !0);
					B(() => K(n, r().note)), G(e, t);
				};
				q(m, (e) => {
					r().note && e(h);
				});
				var g = R(z(m, 2), !0);
				O(a), B((e, t, n) => {
					Ur(o, 1, `pill ${e ?? ""}`, "svelte-t4vask"), K(s, r().state), K(l, t), K(g, n);
				}, [
					() => w(r().state),
					() => te(n(), "fields").step || n().title,
					() => ee(r().at)
				]), G(e, a);
			}, (e) => {
				var t = vr(), n = Qt(t), r = (e) => {
					var t = vr(), n = Qt(t), r = (e) => {
						var t = Cc(), n = L(t), r = R(n, !0), i = R(z(n), !0);
						we(), O(t), B((e) => {
							K(r, U(s).status), K(i, e);
						}, [() => te(U(s), "fields").step || U(s).title]), G(e, t);
					};
					q(n, (e) => {
						U(c) || e(r);
					}), G(e, t);
				}, i = (e) => {
					G(e, wc());
				};
				q(n, (e) => {
					U(s) ? e(r) : e(i, -1);
				}), G(e, t);
			});
			var S = z(x, 2), C = (e) => {
				var n = Ec(), r = z(L(n)), i = L(r), o = (e) => {
					var n = Tc(), r = R(n, !0);
					B(() => {
						Jr(n, "href", "?session=" + a().author), K(r, a().author);
					}), cr("click", n, (e) => {
						e.preventDefault(), t.api.openSession(a().author);
					}), G(e, n);
				}, s = /* @__PURE__ */ ut(() => re(a().author)), c = (e) => {
					var t = _r();
					B(() => K(t, a().author)), G(e, t);
				};
				q(i, (e) => {
					U(s) ? e(o) : e(c, -1);
				}), O(r);
				var l = z(r), u = R(l, !0), d = R(z(l), !0);
				O(n), B((e) => {
					K(u, a().body), K(d, e);
				}, [() => ee(a().created_at)]), G(e, n);
			};
			q(S, (e) => {
				a() && e(C);
			}), O(b), O(l), B((e) => {
				Jr(l, "data-run", r().id), K(f, r().title), K(m, r().id), K(h, e);
			}, [() => ee(r().updated_at)]), cr("click", u, () => t.api.openTicket(r().id)), G(e, l);
		}), O(n), G(e, n);
	}, E = (e) => {
		G(e, kc());
	}, xe = (e) => {
		G(e, Ac());
	};
	q(ye, (e) => {
		U(i).length ? e(be) : U(o) ? e(xe, -1) : e(E, 1);
	}), O(fe), G(e, oe), Ue();
}
lr(["click"]);
// plugins/attention/index.svelte.js
var Nc = new URL("./style.css", import.meta.url).href, Pc = class extends HTMLElement {
	#e = /* @__PURE__ */ N(F({
		api: null,
		context: {},
		config: {}
	}));
	get props() {
		return U(this.#e);
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
		if (!document.querySelector(`link[href="${Nc}"]`)) {
			let e = document.createElement("link");
			e.rel = "stylesheet", e.href = Nc, document.head.append(e);
		}
		this.style.display = "contents", this.app = Tr(Mc, {
			target: this,
			props: this.props
		});
	}
	disconnectedCallback() {
		this.app && kr(this.app), this.app = null;
	}
};
customElements.get("desk-attention-view") || customElements.define("desk-attention-view", Pc);
