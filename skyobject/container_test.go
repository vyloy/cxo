package skyobject

import (
	"testing"

	"github.com/skycoin/cxo/data"
	"github.com/skycoin/skycoin/src/cipher"
)

func TestNewContainer(t *testing.T) {
	t.Run("missing DB", func(t *testing.T) {
		defer shouldPanic(t)
		NewContainer(nil, nil)
	})
	t.Run("registry", func(t *testing.T) {
		reg := NewRegistry()
		db := data.NewMemoryDB()
		c := NewContainer(db, reg)
		if c.DB() == nil {
			t.Error("missing database")
		} else if c.DB() != db {
			t.Error("wrong database")
		} else if c.coreRegistry != reg {
			t.Error("wrong core registry")
		} else {
			if reg.done != true {
				t.Error("missing (*Registry).Done in NewContainer")
			}
			if _, ok := c.DB().Get(cipher.SHA256(reg.Reference())); !ok {
				t.Error("registry wasn't saved")
			}
		}
		if _, err := c.Registry(reg.Reference()); err != nil {
			t.Error("can't give core registry by reference")
		}
	})
}

func TestContainer_AddRegistry(t *testing.T) {
	c := NewContainer(data.NewMemoryDB(), nil)
	reg := NewRegistry()
	c.AddRegistry(reg)
	if reg.done != true {
		t.Error("missing (*Registry).Done in AddRegistry")
	} else if _, err := c.Registry(reg.Reference()); err != nil {
		t.Error("can't give registyr by reference")
	} else if _, ok := c.DB().Get(cipher.SHA256(reg.Reference())); !ok {
		t.Error("registry wasn't saved")
	}
}

func TestContainer_CoreRegistry(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		c := NewContainer(data.NewMemoryDB(), nil)
		if c.CoreRegistry() != nil {
			t.Error("unexpected core registry")
		}
	})
	t.Run("got", func(t *testing.T) {
		reg := NewRegistry()
		c := NewContainer(data.NewMemoryDB(), reg)
		if c.CoreRegistry() != reg {
			t.Error("missing or wrong registry")
		}
	})
}

func TestContainer_Registry(t *testing.T) {
	// core
	cr := NewRegistry()
	cr.Register("cxo.User", User{})
	// add
	ar := NewRegistry()
	ar.Register("cxo.Developer", Developer{})
	//
	c := NewContainer(data.NewMemoryDB(), cr)
	c.AddRegistry(ar)
	//
	if _, err := c.Registry(RegistryReference{}); err == nil {
		t.Error("missing error")
	}
	if _, err := c.Registry(cr.Reference()); err != nil {
		t.Error("missing core registry")
	}
	if _, err := c.Registry(ar.Reference()); err != nil {
		t.Error("missing added registry")
	}
}

func TestContainer_WantRegistry(t *testing.T) {
	t.Run("dont want", func(t *testing.T) {
		c := NewContainer(data.NewMemoryDB(), nil)
		reg := NewRegistry()
		reg.Done()
		if c.WantRegistry(reg.Reference()) {
			t.Error("unexpected want")
		}
	})
	t.Run("want", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register("cxo.User", User{})
		c1 := NewContainer(data.NewMemoryDB(), reg)
		pk, sk := cipher.GenerateKeyPair()
		r, err := c1.NewRoot(pk, sk)
		if err != nil {
			t.Fatal(err)
		}
		_, rp, err := r.Inject("cxo.User", User{"Alice", 20, nil})
		if err != nil {
			t.Fatal(err)
		}
		if c1.WantRegistry(reg.Reference()) {
			t.Error("missing want")
		}
		c2 := NewContainer(data.NewMemoryDB(), nil)
		if _, r := c2.AddRootPack(&rp); r != nil {
			t.Fatal(err)
		}
		if !c2.WantRegistry(reg.Reference()) {
			t.Error("missing want")
		}
	})
}

func TestContainer_Registries(t *testing.T) {
	t.Run("core", func(t *testing.T) {
		reg := NewRegistry()
		c := NewContainer(data.NewMemoryDB(), reg)
		regs := c.Registries()
		if len(regs) != 1 {
			t.Error("wrong registries")
		} else if regs[0] != reg.Reference() {
			t.Error("wrong reference")
		}
	})
	t.Run("no core", func(t *testing.T) {
		c := NewContainer(data.NewMemoryDB(), nil)
		reg := NewRegistry()
		c.AddRegistry(reg)
		regs := c.Registries()
		if len(regs) != 1 {
			t.Error("wrong registries")
		} else if regs[0] != reg.Reference() {
			t.Error("wrong reference")
		}
	})
}

func TestContainer_DB(t *testing.T) {
	db := data.NewMemoryDB()
	c := NewContainer(db, nil)
	if c.DB() != db {
		t.Error("wrong db")
	}
}

func TestContainer_Get(t *testing.T) {
	db := data.NewMemoryDB()
	c := NewContainer(db, nil)
	value := []byte("hey ho")
	key := cipher.SumSHA256(value)
	db.Set(key, value)
	data, ok := c.Get(Reference(key))
	if !ok {
		t.Error("misisng object")
	} else if string(data) != string(value) {
		t.Error("wrong value")
	}
}

func TestContainer_Set(t *testing.T) {
	db := data.NewMemoryDB()
	c := NewContainer(db, nil)
	value := []byte("hey ho")
	key := cipher.SumSHA256(value)
	c.Set(Reference(key), value)
	data, ok := db.Get(key)
	if !ok {
		t.Error("misisng object")
	} else if string(data) != string(value) {
		t.Error("wrong value")
	}
}

func TestContainer_NewRoot(t *testing.T) {
	t.Run("no core", func(t *testing.T) {
		c := NewContainer(data.NewMemoryDB(), nil)
		pk, sk := cipher.GenerateKeyPair()
		if _, err := c.NewRoot(pk, sk); err == nil {
			t.Error("misisng error")
		} else if err != ErrNoCoreRegistry {
			t.Error("wrong error")
		}
	})
	t.Run("invalid pk sk", func(t *testing.T) {
		c := NewContainer(data.NewMemoryDB(), NewRegistry())
		pk, sk := cipher.GenerateKeyPair()
		if _, err := c.NewRoot(cipher.PubKey{}, sk); err == nil {
			t.Error("misisng error")
		}
		if _, err := c.NewRoot(pk, cipher.SecKey{}); err == nil {
			t.Error("misisng error")
		}
	})
	t.Run("first", func(t *testing.T) {
		reg := NewRegistry()
		db := data.NewMemoryDB()
		c := NewContainer(db, reg)
		pk, sk := cipher.GenerateKeyPair()
		r, err := c.NewRoot(pk, sk)
		if err != nil {
			t.Fatal(err)
		}
		if r.RegistryReference() != reg.Reference() {
			t.Error("wrong reg ref")
		}
		if r.Seq() != 0 {
			t.Error("wrong seq")
		}
		if r.IsReadOnly() {
			t.Error("read only")
		}
		if r.IsAttached() {
			t.Error("attached")
		}
		if r.PrevHash() != (RootReference{}) {
			t.Error("wrong prev. hash")
		}
	})
	t.Run("non first", func(t *testing.T) {
		reg := NewRegistry()
		db := data.NewMemoryDB()
		c := NewContainer(db, reg)
		pk, sk := cipher.GenerateKeyPair()
		r, err := c.NewRoot(pk, sk)
		if err != nil {
			t.Fatal(err)
		}
		rp, err := r.Touch()
		if err != nil {
			t.Fatal(err)
		}
		r, err = c.NewRoot(pk, sk)
		if err != nil {
			t.Fatal(err)
		}
		if r.RegistryReference() != reg.Reference() {
			t.Error("wrong reg ref")
		}
		if r.Seq() != 1 {
			t.Error("wrong seq")
		}
		if r.IsReadOnly() {
			t.Error("read only")
		}
		if r.IsAttached() {
			t.Error("attached")
		}
		if r.PrevHash() != RootReference(rp.Hash) {
			t.Error("wrong prev. hash")
		}
	})
}

func TestContainer_NewRootReg(t *testing.T) {
	t.Run("invalid pk sk", func(t *testing.T) {
		reg := NewRegistry()
		reg.Done()
		rr := reg.Reference()
		c := NewContainer(data.NewMemoryDB(), nil)
		pk, sk := cipher.GenerateKeyPair()
		if _, err := c.NewRootReg(cipher.PubKey{}, sk, rr); err == nil {
			t.Error("misisng error")
		}
		if _, err := c.NewRootReg(pk, cipher.SecKey{}, rr); err == nil {
			t.Error("misisng error")
		}
	})
	t.Run("first", func(t *testing.T) {
		reg := NewRegistry()
		reg.Done()
		rr := reg.Reference()
		db := data.NewMemoryDB()
		c := NewContainer(db, nil)
		pk, sk := cipher.GenerateKeyPair()
		r, err := c.NewRootReg(pk, sk, rr)
		if err != nil {
			t.Fatal(err)
		}
		if r.RegistryReference() != rr {
			t.Error("wrong reg ref")
		}
		if r.Seq() != 0 {
			t.Error("wrong seq")
		}
		if r.IsReadOnly() {
			t.Error("read only")
		}
		if r.IsAttached() {
			t.Error("attached")
		}
		if r.PrevHash() != (RootReference{}) {
			t.Error("wrong prev. hash")
		}
	})
	t.Run("non first", func(t *testing.T) {
		reg := NewRegistry()
		reg.Done()
		rr := reg.Reference()
		db := data.NewMemoryDB()
		c := NewContainer(db, nil)
		pk, sk := cipher.GenerateKeyPair()
		r, err := c.NewRootReg(pk, sk, rr)
		if err != nil {
			t.Fatal(err)
		}
		rp, err := r.Touch()
		if err != nil {
			t.Fatal(err)
		}
		r, err = c.NewRootReg(pk, sk, rr)
		if err != nil {
			t.Fatal(err)
		}
		if r.RegistryReference() != rr {
			t.Error("wrong reg ref")
		}
		if r.Seq() != 1 {
			t.Error("wrong seq")
		}
		if r.IsReadOnly() {
			t.Error("read only")
		}
		if r.IsAttached() {
			t.Error("attached")
		}
		if r.PrevHash() != RootReference(rp.Hash) {
			t.Error("wrong prev. hash")
		}
	})
}

func isRefsEqual(a, b []Dynamic) (equal bool) {
	if len(a) != len(b) {
		return // false
	}
	for i, r := range a {
		if r != b[i] {
			return // false
		}
	}
	equal = true
	return
}

func TestContainer_AddRootPack(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		// create
		c := NewContainer(data.NewMemoryDB(), NewRegistry())
		pk, sk := cipher.GenerateKeyPair()
		cr, err := c.NewRoot(pk, sk)
		if err != nil {
			t.Fatal(err)
		}
		rp, err := cr.Touch()
		if err != nil {
			t.Fatal(err)
		}
		// add
		a := NewContainer(data.NewMemoryDB(), nil)
		ar, err := a.AddRootPack(&rp)
		if err != nil {
			t.Fatal(err)
		}
		if ar.Seq() != cr.Seq() {
			t.Error("wrong seq")
		}
		if ar.Hash() != cr.Hash() {
			t.Error("wrong hash")
		}
		if ar.PrevHash() != cr.PrevHash() {
			t.Error("wrong prev reference")
		}
		if ar.RegistryReference() != cr.RegistryReference() {
			t.Error("wrong reg. ref")
		}
		if !isRefsEqual(ar.Refs(), cr.Refs()) {
			t.Error("wrong refs")
		}
		if !ar.IsAttached() {
			t.Error("not attached")
		}
		if !ar.IsReadOnly() {
			t.Error("can edit")
		}
	})
	t.Run("deserialize error", func(t *testing.T) {
		var rp data.RootPack
		_, err := NewContainer(data.NewMemoryDB(), nil).AddRootPack(&rp)
		if err == nil {
			t.Error("missing error")
		}
	})
	t.Run("invalid signature", func(t *testing.T) {
		// create
		c := NewContainer(data.NewMemoryDB(), NewRegistry())
		pk, sk := cipher.GenerateKeyPair()
		cr, err := c.NewRoot(pk, sk)
		if err != nil {
			t.Fatal(err)
		}
		rp, err := cr.Touch()
		if err != nil {
			t.Fatal(err)
		}
		wrongHashForSign := cipher.SumSHA256([]byte("hey ho"))
		rp.Sig = cipher.SignHash(wrongHashForSign, sk)
		// add
		a := NewContainer(data.NewMemoryDB(), nil)
		_, err = a.AddRootPack(&rp)
		if err == nil {
			t.Error("missing error")
		}
	})
	t.Run("wrong hash", func(t *testing.T) {
		// create
		c := NewContainer(data.NewMemoryDB(), NewRegistry())
		pk, sk := cipher.GenerateKeyPair()
		cr, err := c.NewRoot(pk, sk)
		if err != nil {
			t.Fatal(err)
		}
		rp, err := cr.Touch()
		if err != nil {
			t.Fatal(err)
		}
		rp.Hash = cipher.SumSHA256([]byte("hey ho"))
		// add
		a := NewContainer(data.NewMemoryDB(), nil)
		_, err = a.AddRootPack(&rp)
		if err == nil {
			t.Error("missing error")
		}
	})
}

func TestContainer_LastRoot(t *testing.T) {
	//
}

func TestContainer_LastRootSk(t *testing.T) {
	//
}

func TestContainer_LastFullRoot(t *testing.T) {
	c := getCont()
	pk, sk := cipher.GenerateKeyPair()
	r, err := c.NewRoot(pk, sk)
	if err != nil {
		t.Error(err)
		return
	}
	r.Inject("cxo.User", User{"Alice", 15, nil})
	r.Inject("cxo.User", User{"Eva", 16, nil})
	r.Inject("cxo.User", User{"Ammy", 17, nil})
	full := c.LastFullRoot(pk)
	if full == nil {
		t.Error("misisng last full root")
	}
}

func TestContainer_Feeds(t *testing.T) {
	//
}

func TestContainer_WantFeed(t *testing.T) {
	//
}

func TestContainer_GotFeed(t *testing.T) {
	//
}

func TestContainer_DelFeed(t *testing.T) {
	//
}

func TestContainer_DelRootsBefore(t *testing.T) {
	//
}

func TestContainer_GC(t *testing.T) {
	//
}
