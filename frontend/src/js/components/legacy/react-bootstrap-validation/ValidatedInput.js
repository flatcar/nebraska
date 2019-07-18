import PropTypes from 'prop-types';
import React from 'react';
import { Input } from '../react-bootstrap';

export default class ValidatedInput extends Input {
    constructor(props) {
        super(props);

        if (!props._registerInput || !props._unregisterInput) {
            throw new Error('Input must be placed inside the Form component');
        }
    }

    componentWillMount() {
        if (Input.prototype.componentWillMount) {
            super.componentWillMount();
        }

        this.props._registerInput(this);
    }

    componentWillUnmount() {
        if (Input.prototype.componentWillUnmount) {
            super.componentWillUnmount();
        }

        this.props._unregisterInput(this);
    }
}

ValidatedInput.propTypes = Object.assign({}, Input.propTypes, {
    name           : PropTypes.string.isRequired,
    validationEvent: PropTypes.oneOf([
        '', 'onChange', 'onBlur', 'onFocus'
    ]),
    validate       : PropTypes.oneOfType([
        PropTypes.func,
        PropTypes.string
    ]),
    errorHelp      : PropTypes.oneOfType([
        PropTypes.string,
        PropTypes.object
    ])
});

ValidatedInput.defaultProps = Object.assign({}, Input.defaultProps, {
    validationEvent: ''
});
