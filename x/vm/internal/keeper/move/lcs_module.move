address 0x123 {
	module Foo {
        use 0x1::Vector;

        struct Inner {
            a: u8,
            b: bool
        }

	    resource struct Bar {
	        u8Val:    u8,
            u64Val:   u64,
            u128Val:  u128,
            boolVal:  bool,
            addrVal:  address,
            vU8Val:   vector<u8>,
            vU64Val:  vector<u64>,
            inStruct: Inner,
            vComplex: vector<Inner>
        }

        public fun BuildAndStoreRes(sender: &signer) {
            let vU8 = Vector::empty<u8>();
            Vector::push_back(&mut vU8, 100);
            Vector::push_back(&mut vU8, 200);

            let vU64 = Vector::empty<u64>();
            Vector::push_back(&mut vU64, 1);
            Vector::push_back(&mut vU64, 2);

            let inSt = Inner {
                a: 128,
                b: false
            };

            let vClx = Vector::empty<Inner>();
            Vector::push_back(&mut vClx, Inner {
                a: 1,
                b: false
            });
            Vector::push_back(&mut vClx, Inner {
                a: 2,
                b: true
            });

        	let res = Bar {
        	    u8Val:    100,
                u64Val:   10000,
                u128Val:  12345678910111213141516171819,
                boolVal:  true,
                addrVal:  0x1::Signer::address_of(sender),
                vU8Val:   vU8,
                vU64Val:  vU64,
                inStruct: inSt,
                vComplex: vClx
            };
            move_to<Bar>(sender, res);
        }
	}
}