import PropTypes from 'prop-types';
import React from 'react';
import InputContainer from './InputContainer';
import ValidatedInput from './ValidatedInput';
import RadioGroup from './RadioGroup';

import Validator from './Validator';
import FileValidator from './FileValidator';

function getInputErrorMessage(input, ruleName) {
    let errorHelp = input.props.errorHelp;

    if (typeof errorHelp === 'object') {
        return errorHelp[ruleName];
    } else {
        return errorHelp;
    }
}

export default class Form extends InputContainer {
    constructor(props) {
        super(props);

        this.state = {
            isValid: true,
            invalidInputs: {}
        };
    }

    componentWillMount() {
        super.componentWillMount();

        this._validators = {};
    }

    registerInput(input) {
        super.registerInput(input);

        if (typeof input.props.validate === 'string') {
            this._validators[input.props.name] = this._compileValidationRules(input, input.props.validate);
        }
    }

    unregisterInput(input) {
        super.unregisterInput(input);

        delete this._validators[input.props.name];
    }

    render() {
        return (
            <form ref="form"
                  onSubmit={this._handleSubmit.bind(this)}
                  method={this.props.method}
                  action="#"
                  className={this.props.className}>
                {this._renderChildren(this.props.children)}
            </form>
        );
    }

    getValues() {
        return Object.keys(this._inputs).reduce((values, name) => {
            values[name] = this._getValue(name);

            return values;
        }, {});
    }

    submit() {
        this._handleSubmit();
    }

    _renderChildren(children) {
        if (typeof children !== 'object' || children === null) {
            return children;
        }

        let childrenCount = React.Children.count(children);

        if (childrenCount > 1) {
            return React.Children.map(children, child => this._renderChild(child));
        } else if (childrenCount === 1) {
            return this._renderChild(Array.isArray(children) ? children[0] : children);
        }
    }

    _renderChild(child) {
        if (typeof child !== 'object' || child === null) {
            return child;
        }

        let model = this.props.model || {};

        if (child.type === ValidatedInput ||
            child.type === RadioGroup || (
                child.type &&
                child.type.prototype !== null && (
                    child.type.prototype instanceof ValidatedInput ||
                    child.type.prototype instanceof RadioGroup
                )
            )
        ) {
            let name = child.props && child.props.name;

            if (!name) {
                throw new Error('Can not add input without "name" attribute');
            }

            let newProps = {
                _registerInput  : this.registerInput.bind(this),
                _unregisterInput: this.unregisterInput.bind(this)
            };

            let evtName = child.props.validationEvent ?
                    child.props.validationEvent : this.props.validationEvent;

            let origCallback = child.props[evtName];

            newProps[evtName] = e => {
                this._validateInput(name);

                return origCallback && origCallback(e);
            };

            if (name in model) {
                if (child.props.type === 'checkbox') {
                    newProps.defaultChecked = model[name];
                } else {
                    newProps.defaultValue = model[name];
                }
            }

            let error = this._hasError(name);

            if (error) {
                newProps.bsStyle = 'error';

                if (typeof error === 'string') {
                    newProps.help = error;
                } else if (child.props.errorHelp) {
                    newProps.help = <React.Fragment>child.props.errorHelp</React.Fragment>;
                }
            }

            return React.cloneElement(child, newProps);
        } else {
            return React.cloneElement(child, {}, this._renderChildren(child.props && child.props.children));
        }
    }

    _validateInput(name) {
        this._validateOne(name, this.getValues());
    }

    _hasError(iptName) {
        return this.state.invalidInputs[iptName];
    }

    _setError(iptName, isError, errText) {
        if (isError && errText &&
            typeof errText !== 'string' &&
            typeof errText !== 'boolean')
        {
            errText = errText + '';
        }

        // set value to either bool or error description string
        this.setState({
            invalidInputs: Object.assign(
                this.state.invalidInputs,
                {
                    [iptName]: isError ? errText || true : false
                }
            )
        });
    }

    _validateOne(iptName, context) {
        let input = this._inputs[iptName];

        if (Array.isArray(input)) {
            console.warn('Multiple inputs use the same name "' + iptName + '"');

            return false;
        }

        let value = context[iptName];
        let isValid = true;
        let validate = input.props.validate;
        let result, error;

         if (typeof validate === 'function') {
            result = validate(value, context);
        } else if (typeof validate === 'string') {
            result = this._validators[iptName](value);
        } else {
            result = true;
        }

		if (typeof this.props.validateOne === 'function') {
            result = this.props.validateOne(iptName, value, context, result);
        } 
        // if result is !== true, it is considered an error
        // it can be either bool or string error
        if (result !== true) {
            isValid = false;

            if (typeof result === 'string') {
                error = result;
            }
        }

        this._setError(iptName, !isValid, error);

        return isValid;
    }

    _validateAll(context) {
        let isValid = true;
        let errors = [];

        if (typeof this.props.validateAll === 'function') {
            let result = this.props.validateAll(context);

            if (result !== true) {
                isValid = false;

                Object.keys(result).forEach(iptName => {
                    errors.push(iptName);

                    this._setError(iptName, true, result[iptName]);
                });
            }
        } else {
            Object.keys(this._inputs).forEach(iptName => {
                if (!this._validateOne(iptName, context)) {
                    isValid = false;
                    errors.push(iptName);
                }
            });
        }

        return {
            isValid: isValid,
            errors: errors
        };
    }

    _compileValidationRules(input, ruleProp) {
        let rules = ruleProp.split(',').map(rule => {
            let params = rule.split(':');
            let name = params.shift();
            let inverse = name[0] === '!';

            if (inverse) {
                name = name.substr(1);
            }

            return { name, inverse, params };
        });

        let validator = (input.props && input.props.type) === 'file' ? FileValidator : Validator;

        return val => {
            let result = true;

            rules.forEach(rule => {
                if (typeof validator[rule.name] !== 'function') {
                    throw new Error('Invalid input validation rule "' + rule.name + '"');
                }

                let ruleResult = validator[rule.name](val, ...rule.params);

                if (rule.inverse) {
                    ruleResult = !ruleResult;
                }

                if (result === true && ruleResult !== true) {
                    result = getInputErrorMessage(input, rule.name) ||
                        getInputErrorMessage(this, rule.name) || false;
                }
            });

            return result;
        };
    }

    _getValue(iptName) {
        let input = this._inputs[iptName];

        if (Array.isArray(input)) {
            console.warn('Multiple inputs use the same name "' + iptName + '"');

            return false;
        }

        let value;

        if (input.props.type === 'checkbox') {
            value = input.getChecked();
        } else if (input.props.type === 'file') {
            value = input.getInputDOMNode().files;
        } else {
            value = input.getValue();
        }

        return value;
    }

    _handleSubmit(e) {
        if (e) {
            e.preventDefault();
        }

        let values = this.getValues();

        let { isValid, errors } = this._validateAll(values);

        if (isValid) {
            this.props.onValidSubmit(values);
        } else {
            this.props.onInvalidSubmit(errors, values);
        }
    }
}

Form.propTypes = {
    className      : PropTypes.string,
    model          : PropTypes.object,
    method         : PropTypes.oneOf(['get', 'post']),
    onValidSubmit  : PropTypes.func.isRequired,
    onInvalidSubmit: PropTypes.func,
    validateOne    : PropTypes.func,
    validateAll    : PropTypes.func,
    validationEvent: PropTypes.oneOf([
        'onChange', 'onBlur', 'onFocus'
    ]),
    errorHelp      : PropTypes.oneOfType([
        PropTypes.string,
        PropTypes.object
    ])
};

Form.defaultProps = {
    model          : {},
    validationEvent: 'onChange',
    method         : 'get',
    onInvalidSubmit: () => {}
};
