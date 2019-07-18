import validator from 'validator';

/**
 * Returns true if the value is not empty
 *
 * @params {String} val
 * @returns {Boolean}
 */
validator.required = function(val) {
    return !validator.isEmpty(val);
}

/**
 * Returns true if the value is boolean true
 *
 * @params {String} val
 * @returns {Boolean}
 */
validator.isChecked = function(val) {
    // compare it against string representation of a bool value, because
    // validator ensures all incoming values are coerced to strings
    // https://github.com/chriso/validator.js#strings-only
    return val === 'true';
}

export default validator;
