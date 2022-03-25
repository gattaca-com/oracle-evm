pragma solidity >=0.8.0;

interface NativePriceOracleInterface {
    // Set [addr] to have the admin role over the minter list
    function getPrice(uint256 identifier) external view returns (uint256);

    function setPrice(uint256 identifier, uint256 value) external;
}
