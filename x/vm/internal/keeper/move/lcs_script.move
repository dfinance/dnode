script {
	use 0x123::Foo;

	fun main(sender: &signer) {
	    Foo::BuildAndStoreRes(sender)
	}
}