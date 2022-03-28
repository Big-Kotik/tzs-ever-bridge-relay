pragma ton-solidity >= 0.35.0;
pragma AbiHeader expire;
pragma AbiHeader pubkey;


contract test {
    uint public balance;

    event UnwrapTokenEvent(string addr, uint amount);

    constructor() public {
        require(tvm.pubkey() != 0, 101);
        require(msg.pubkey() == tvm.pubkey(), 102);
        tvm.accept();
        balance = 0;
    }

    modifier checkOwnerAndAccept {
        require(msg.pubkey() == tvm.pubkey(), 102);
        tvm.accept();
        _;
    }

    function unwrapToken(string addr, uint256 amount) public checkOwnerAndAccept {
        tvm.accept();
        balance += amount;
        emit UnwrapTokenEvent(addr, amount);
    }
}